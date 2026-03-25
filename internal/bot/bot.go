package bot

import (
	"log"
	"musicbot/internal/musicapi"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	cfg Config
	dg  *discordgo.Session
	api *musicapi.Client

	pm *PlaybackManager
}

func New(cfg Config) (*Bot, error) {
	dg, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return nil, err
	}

	b := &Bot{
		cfg: cfg,
		dg:  dg,
		api: musicapi.New(cfg.MusicAPIBase, cfg.MusicPrefix),
	}
	b.pm = NewPlaybackManager(b)

	return b, nil
}

func (b *Bot) Start() error {
	b.dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates
	b.dg.AddHandler(b.onReady)
	b.dg.AddHandler(b.onInteractionCreate)

	var err error
	for i := 0; i < 5; i++ {
		err = b.dg.Open()
		if err == nil {
			return nil
		}
		log.Printf("Failed to connect (attempt %d/5): %v", i+1, err)
		time.Sleep(time.Duration(i+1) * 5 * time.Second)
	}
	return err
}

func (b *Bot) Close() error {
	b.pm.StopAll()
	return b.dg.Close()
}

func (b *Bot) onReady(s *discordgo.Session, r *discordgo.Ready) {
	log.Printf("Logged in as %s", s.State.User.String())
}
