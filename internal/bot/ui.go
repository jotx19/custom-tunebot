package bot

import (
	"fmt"
	"strings"

	"musicbot/internal/musicapi"

	"github.com/bwmarrin/discordgo"
)

// Modern UI color (Discord blurple)
const uiColor = 0x5865F2

type UIState struct {
	Status      string
	VoiceChanID string
	RequestedBy string
}

func NowPlayingEmbed(d *musicapi.SongDetail, ui UIState) *discordgo.MessageEmbed {
	title := strings.TrimSpace(d.Title)
	if title == "" {
		title = "Unknown Title"
	}

	artist := strings.TrimSpace(d.Artist)
	if artist == "" {
		artist = "Unknown Artist"
	}

	status := strings.TrimSpace(ui.Status)
	if status == "" {
		status = "Playing"
	}

	desc := fmt.Sprintf("**%s**\n\n**Status:** `%s`", artist, status)

	embed := &discordgo.MessageEmbed{
		Title:       "🎶 Now Playing",
		Description: desc,
		Color:       uiColor,
		URL:         d.Link,
	}

	if d.Image != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: d.Image}
	}

	voice := mentionChannel(ui.VoiceChanID)

	req := strings.TrimSpace(ui.RequestedBy)
	if req == "" {
		req = "`unknown`"
	}

	embed.Fields = []*discordgo.MessageEmbedField{
		{
			Name:   "Track",
			Value:  fmt.Sprintf("**%s**", title),
			Inline: false,
		},
		{
			Name:   "Voice",
			Value:  voice,
			Inline: true,
		},
		{
			Name:   "Requested by",
			Value:  req,
			Inline: true,
		},
	}

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Pause/Resume toggles • Stop ends playback • Leave disconnects",
	}

	return embed
}

// PlayerControls returns modern controls in two rows (NO Open button):
// Row 1: Toggle (Pause/Resume) + Stop
// Row 2: Leave
func PlayerControls(isPaused bool) []discordgo.MessageComponent {
	// Toggle button (Pause ↔ Resume)
	toggleLabel := "Pause"
	toggleEmoji := "⏸️"
	toggleID := ctrlPauseID
	toggleStyle := discordgo.PrimaryButton

	if isPaused {
		toggleLabel = "Resume"
		toggleEmoji = "▶️"
		toggleID = ctrlResumeID
		toggleStyle = discordgo.SuccessButton
	}

	row1 := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				CustomID: toggleID,
				Label:    toggleLabel,
				Style:    toggleStyle,
				Emoji:    &discordgo.ComponentEmoji{Name: toggleEmoji},
			},
			discordgo.Button{
				CustomID: ctrlStopID,
				Label:    "Stop",
				Style:    discordgo.DangerButton,
				Emoji:    &discordgo.ComponentEmoji{Name: "⏹️"},
			},
		},
	}

	row2 := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				CustomID: ctrlLeaveID,
				Label:    "Leave",
				Style:    discordgo.SecondaryButton,
				Emoji:    &discordgo.ComponentEmoji{Name: "🚪"},
			},
		},
	}

	return []discordgo.MessageComponent{row1, row2}
}

func mentionChannel(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return "`unknown`"
	}
	return "<#" + id + ">"
}
