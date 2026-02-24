package bot

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"

	"layeh.com/gopus"
)

const (
	sampleRate   = 48000
	channels     = 2
	frameSize    = 960  // 20ms @ 48kHz
	maxOpusBytes = 4000 // max packet size
)

func (b *Bot) userVoiceChannelID(guildID, userID string) (string, error) {
	// Try cache first
	g, err := b.dg.State.Guild(guildID)
	if err == nil {
		for _, vs := range g.VoiceStates {
			if vs.UserID == userID && vs.ChannelID != "" {
				return vs.ChannelID, nil
			}
		}
	}

	// Fallback: fetch from API
	g, err = b.dg.Guild(guildID)
	if err != nil {
		return "", err
	}
	for _, vs := range g.VoiceStates {
		if vs.UserID == userID && vs.ChannelID != "" {
			return vs.ChannelID, nil
		}
	}
	return "", errors.New("user not in a voice channel")
}

func (b *Bot) playURLWithPause(ctx context.Context, p *Player, audioURL string) error {
	// Give discord voice connection a moment to be ready
	time.Sleep(300 * time.Millisecond)

	// ffmpeg: decode URL -> raw PCM s16le 48k stereo -> stdout
	ff := exec.Command(
		b.cfg.FFmpegPath,
		"-reconnect", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "5",
		"-i", audioURL,
		"-f", "s16le",
		"-ar", "48000",
		"-ac", "2",
		"pipe:1",
	)

	stdout, err := ff.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, _ := ff.StderrPipe()

	if err := ff.Start(); err != nil {
		return err
	}
	defer func() { _ = ff.Process.Kill() }()

	// Drain stderr so ffmpeg never blocks (important!)
	go func() { _, _ = io.Copy(io.Discard, stderr) }()

	enc, err := gopus.NewEncoder(sampleRate, channels, gopus.Audio)
	if err != nil {
		return fmt.Errorf("opus encoder: %w", err)
	}

	vc := p.vc
	_ = vc.Speaking(true)
	defer func() { _ = vc.Speaking(false) }()

	reader := bufio.NewReaderSize(stdout, 1<<20)

	pcmFrame := make([]int16, frameSize*channels)

	for {
		// Stop?
		select {
		case <-ctx.Done():
			return errors.New("stopped")
		default:
		}

		// Pause support (blocks here while paused)
		if err := p.waitIfPaused(ctx); err != nil {
			return err
		}

		// Read 20ms PCM frame
		if err := readInt16Frame(reader, pcmFrame); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("read pcm: %w", err)
		}

		// gopus Encode returns []byte packet (NOT int)
		packet, err := enc.Encode(pcmFrame, frameSize, maxOpusBytes)
		if err != nil {
			return fmt.Errorf("opus encode: %w", err)
		}

		// Send opus packet to Discord
		select {
		case vc.OpusSend <- packet:
		case <-ctx.Done():
			return errors.New("stopped")
		case <-time.After(2 * time.Second):
			return errors.New("opus send timeout (voice not ready)")
		}
	}

	_ = ff.Wait()
	return nil
}

func readInt16Frame(r *bufio.Reader, dst []int16) error {
	// dst length is samples * channels
	buf := make([]byte, len(dst)*2)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return err
	}
	for i := 0; i < len(dst); i++ {
		dst[i] = int16(binary.LittleEndian.Uint16(buf[i*2 : i*2+2]))
	}
	return nil
}
