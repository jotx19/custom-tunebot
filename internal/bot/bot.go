package bot

import (
	"log"

	"musicbot/internal/musicapi"

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
	// Need VoiceStates to know which VC the user is in
	b.dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates

	b.dg.AddHandler(b.onReady)
	b.dg.AddHandler(b.onInteractionCreate)

	if err := b.dg.Open(); err != nil {
		return err
	}

	return b.registerCommands()
}

func (b *Bot) Close() error {
	b.pm.StopAll()
	return b.dg.Close()
}

func (b *Bot) onReady(s *discordgo.Session, r *discordgo.Ready) {
	log.Printf("Logged in as %s", s.State.User.String())
}
