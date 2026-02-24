package bot

import "github.com/bwmarrin/discordgo"

func (b *Bot) registerCommands() error {
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

	appID := b.dg.State.User.ID

	if b.cfg.GuildID != "" {
		for _, c := range cmds {
			if _, err := b.dg.ApplicationCommandCreate(appID, b.cfg.GuildID, c); err != nil {
				return err
			}
		}
		return nil
	}

	for _, c := range cmds {
		if _, err := b.dg.ApplicationCommandCreate(appID, "", c); err != nil {
			return err
		}
	}
	return nil
}
