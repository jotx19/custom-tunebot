package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const (
	playSelectID = "play_select_song"

	ctrlPauseID  = "ctrl_pause"
	ctrlResumeID = "ctrl_resume"
	ctrlStopID   = "ctrl_stop"
	ctrlLeaveID  = "ctrl_leave"
)

func (b *Bot) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {

	case discordgo.InteractionApplicationCommand:
		if i.ApplicationCommandData().Name == "play" {
			b.handlePlay(s, i)
		}

	case discordgo.InteractionMessageComponent:
		cid := i.MessageComponentData().CustomID

		switch cid {
		case playSelectID:
			b.handlePickSong(s, i)
		case ctrlPauseID:
			b.handleControl(s, i, "pause")
		case ctrlResumeID:
			b.handleControl(s, i, "resume")
		case ctrlStopID:
			b.handleControl(s, i, "stop")
		case ctrlLeaveID:
			b.handleControl(s, i, "leave")
		}
	}
}

func (b *Bot) handlePlay(s *discordgo.Session, i *discordgo.InteractionCreate) {
	query := strings.TrimSpace(i.ApplicationCommandData().Options[0].StringValue())
	if query == "" {
		replyText(s, i, "Give me a song name or artist.")
		return
	}

	// Ack quickly
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	results, err := b.api.SearchSongs(query)
	if err != nil {
		editReplyText(s, i, "API error: "+err.Error())
		return
	}
	if len(results) == 0 {
		editReplyText(s, i, "No results found.")
		return
	}
	if len(results) > 25 {
		results = results[:25]
	}

	opts := make([]discordgo.SelectMenuOption, 0, len(results))
	for _, song := range results {
		label := truncate(fmt.Sprintf("%s — %s", song.Title, song.Artist), 100)
		desc := truncate(song.Artist, 100)
		opts = append(opts, discordgo.SelectMenuOption{
			Label:       label,
			Description: desc,
			Value:       song.ID,
		})
	}

	menu := discordgo.SelectMenu{
		CustomID:    playSelectID,
		Placeholder: "Pick a track to play…",
		Options:     opts,
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Search Results",
		Description: fmt.Sprintf("Query: **%s**\nSelect a track below.", query),
		Color:       uiColor,
	}

	_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
		Components: &[]discordgo.MessageComponent{
			discordgo.ActionsRow{Components: []discordgo.MessageComponent{menu}},
		},
	})
}

func (b *Bot) handlePickSong(s *discordgo.Session, i *discordgo.InteractionCreate) {
	values := i.MessageComponentData().Values
	if len(values) == 0 {
		replyText(s, i, "No song selected.")
		return
	}
	id := values[0]

	// Remove dropdown immediately (clean)
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{Title: "Loading…", Description: "Fetching track details…", Color: uiColor},
			},
			Components: []discordgo.MessageComponent{},
		},
	})

	detail, err := b.api.GetSongByID(id)
	if err != nil {
		followupText(s, i, "Couldn’t load song details: "+err.Error())
		return
	}

	guildID := i.GuildID
	userID := i.Member.User.ID

	vcID, err := b.userVoiceChannelID(guildID, userID)
	if err != nil || vcID == "" {
		followupText(s, i, "Join a **voice channel** first, then use `/play` again.")
		return
	}

	// Pick a playable URL
	stream := detail.StreamURL
	if stream == "" {
		stream = detail.Link
	}
	if stream == "" {
		followupText(s, i, "No playable audio URL found for this track.")
		return
	}

	requestedBy := "@" + i.Member.User.Username

	// Start playback FIRST (so controls actually work)
	if err := b.pm.Start(guildID, vcID, stream, detail, requestedBy); err != nil {
		followupText(s, i, "Playback error: "+err.Error())
		return
	}

	// Send modern player UI (embed + controls)
	embed := NowPlayingEmbed(detail, UIState{
		Status:      "Playing",
		VoiceChanID: vcID,
		RequestedBy: requestedBy,
	})

	components := PlayerControls(false, detail.Link)

	_, _ = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
}

func (b *Bot) handleControl(s *discordgo.Session, i *discordgo.InteractionCreate, action string) {
	// Ack fast
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	guildID := i.GuildID

	switch action {
	case "pause":
		b.pm.Pause(guildID)
	case "resume":
		b.pm.Resume(guildID)
	case "stop":
		b.pm.Stop(guildID)
	case "leave":
		b.pm.Stop(guildID)
		b.pm.Leave(guildID)
	}

	// Update the message UI to reflect status + toggle button
	track, requestedBy, vcID, ok := b.pm.TrackInfo(guildID)
	if !ok || track == nil {
		// If stopped/left, show a clean "stopped" state
		stopped := &discordgo.MessageEmbed{
			Title:       "Player",
			Description: "**Status:** `Stopped`",
			Color:       uiColor,
		}
		_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds:     &[]*discordgo.MessageEmbed{stopped},
			Components: &[]discordgo.MessageComponent{},
		})
		return
	}

	paused := b.pm.IsPaused(guildID)
	status := "Playing"
	if paused {
		status = "Paused"
	}

	embed := NowPlayingEmbed(track, UIState{
		Status:      status,
		VoiceChanID: vcID,
		RequestedBy: requestedBy,
	})

	comps := PlayerControls(paused, track.Link)

	_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &comps,
	})
}
