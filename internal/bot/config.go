package bot

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	Token        string
	GuildID      string
	MusicAPIBase string
	MusicPrefix  string
	FFmpegPath   string
}

func LoadConfigFromEnv() (Config, error) {
	token := strings.TrimSpace(os.Getenv("DISCORD_TOKEN"))
	if token == "" {
		return Config{}, errors.New("DISCORD_TOKEN missing")
	}

	base := strings.TrimSpace(os.Getenv("MUSIC_API_BASE"))
	if base == "" {
		return Config{}, errors.New("MUSIC_API_BASE missing")
	}
	base = strings.TrimRight(base, "/")

	prefix := strings.TrimSpace(os.Getenv("MUSIC_API_PREFIX"))
	if prefix == "" {
		prefix = "/api"
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	ff := strings.TrimSpace(os.Getenv("FFMPEG_PATH"))
	if ff == "" {
		ff = "ffmpeg"
	}

	return Config{
		Token:        token,
		GuildID:      strings.TrimSpace(os.Getenv("GUILD_ID")),
		MusicAPIBase: base,
		MusicPrefix:  prefix,
		FFmpegPath:   ff,
	}, nil
}
