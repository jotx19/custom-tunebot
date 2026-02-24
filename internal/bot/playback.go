package bot

import (
	"context"
	"errors"
	"sync"

	"musicbot/internal/musicapi"

	"github.com/bwmarrin/discordgo"
)

type PlaybackManager struct {
	bot *Bot

	mu      sync.Mutex
	players map[string]*Player // guildID -> player
}

func NewPlaybackManager(b *Bot) *PlaybackManager {
	return &PlaybackManager{
		bot:     b,
		players: make(map[string]*Player),
	}
}

type Player struct {
	guildID string
	vcID    string
	vc      *discordgo.VoiceConnection

	track       *musicapi.SongDetail
	requestedBy string

	cancel context.CancelFunc

	mu     sync.Mutex
	paused bool
	cond   *sync.Cond
}

func (pm *PlaybackManager) Start(guildID, vcID, url string, track *musicapi.SongDetail, requestedBy string) error {
	pm.mu.Lock()
	// stop existing
	if p := pm.players[guildID]; p != nil {
		p.stop()
		delete(pm.players, guildID)
	}
	pm.mu.Unlock()

	vc, err := pm.bot.dg.ChannelVoiceJoin(guildID, vcID, false, true)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	p := &Player{
		guildID:     guildID,
		vcID:        vcID,
		vc:          vc,
		cancel:      cancel,
		track:       track,
		requestedBy: requestedBy,
	}
	p.cond = sync.NewCond(&p.mu)

	pm.mu.Lock()
	pm.players[guildID] = p
	pm.mu.Unlock()

	// playback goroutine
	go func() {
		_ = pm.bot.playURLWithPause(ctx, p, url)
		_ = vc.Disconnect()

		pm.mu.Lock()
		if pm.players[guildID] == p {
			delete(pm.players, guildID)
		}
		pm.mu.Unlock()
	}()

	return nil
}

func (pm *PlaybackManager) Pause(guildID string) {
	if p := pm.get(guildID); p != nil {
		p.mu.Lock()
		p.paused = true
		p.mu.Unlock()
	}
}

func (pm *PlaybackManager) Resume(guildID string) {
	if p := pm.get(guildID); p != nil {
		p.mu.Lock()
		p.paused = false
		p.mu.Unlock()
		p.cond.Broadcast()
	}
}

func (pm *PlaybackManager) Stop(guildID string) {
	if p := pm.get(guildID); p != nil {
		p.stop()
	}
}

func (pm *PlaybackManager) Leave(guildID string) {
	if p := pm.get(guildID); p != nil && p.vc != nil {
		_ = p.vc.Disconnect()
	}
}

func (pm *PlaybackManager) StopAll() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for _, p := range pm.players {
		p.stop()
		if p.vc != nil {
			_ = p.vc.Disconnect()
		}
	}
	pm.players = make(map[string]*Player)
}

func (pm *PlaybackManager) IsPaused(guildID string) bool {
	if p := pm.get(guildID); p != nil {
		p.mu.Lock()
		defer p.mu.Unlock()
		return p.paused
	}
	return false
}

func (pm *PlaybackManager) TrackInfo(guildID string) (track *musicapi.SongDetail, requestedBy string, vcID string, ok bool) {
	if p := pm.get(guildID); p != nil {
		p.mu.Lock()
		defer p.mu.Unlock()
		return p.track, p.requestedBy, p.vcID, p.track != nil
	}
	return nil, "", "", false
}

func (pm *PlaybackManager) get(guildID string) *Player {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.players[guildID]
}

func (p *Player) stop() {
	if p.cancel != nil {
		p.cancel()
	}
}

func (p *Player) waitIfPaused(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for p.paused {
		select {
		case <-ctx.Done():
			return errors.New("stopped")
		default:
		}
		p.cond.Wait()
	}
	return nil
}
