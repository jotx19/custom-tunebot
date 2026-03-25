package bot

import "github.com/bwmarrin/discordgo"

func RegisterCommands(cfg Config) error {
	dg, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return err
	}
	if err := dg.Open(); err != nil {
		return err
	}
	defer dg.Close()

	cmds := []*discordgo.ApplicationCommand{
		{
			Name:        "play",
			Description: "Search songs, pick one from dropdown, then play in your voice channel",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "query",
					Description: "Song name or artist",
					Required:    true,
				},
			},
		},
	}

	appID := dg.State.User.ID
	for _, c := range cmds {
		if _, err := dg.ApplicationCommandCreate(appID, cfg.GuildID, c); err != nil {
			return err
		}
	}
	return nil
}
