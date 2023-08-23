package ui

import (
	"yamusic/api"

	"github.com/charmbracelet/lipgloss"
)

func (m model) renderMainPage() string {
	var tracker string = "\n"
	sidePanel := sideBoxStyle.Render(m.playlistList.View())

	var currentTrack api.Track
	if len(m.playQueue) > 0 {
		currentTrack = m.playQueue[m.currentTrackIdx]
	}

	var playButton string
	isPlaying := m.player != nil && m.player.IsPlaying()
	if isPlaying {
		playButton = activeButtonStyle.Padding(0, 1).Margin(0, 0, 0, 1).Render("■")
	} else {
		playButton = activeButtonStyle.Padding(0, 1).Margin(0, 0, 0, 1).Render("▶")
	}

	trackTitle := trackTitleStyle.Render(currentTrack.Title)
	trackVersion := trackVersionStyle.Render(currentTrack.Version)
	var artists string
	for _, a := range currentTrack.Artists {
		artists += a.Name + ", "
	}
	if len(artists) > 2 {
		artists = artists[:len(artists)-2]
	}
	trackArtist := trackArtistStyle.Render(artists)

	tracker = lipgloss.JoinHorizontal(lipgloss.Top, trackTitle, trackVersion)
	tracker = lipgloss.JoinVertical(lipgloss.Left, tracker, trackArtist)
	tracker = lipgloss.JoinHorizontal(lipgloss.Top, playButton, tracker)
	tracker = lipgloss.JoinVertical(lipgloss.Left, trackProgressStyle.Render(m.trackProgress.View()), tracker)

	tracker = trackBoxStyle.Width(m.width - m.playlistList.Width()).Render(tracker)
	tracker = lipgloss.JoinVertical(lipgloss.Left, trackBoxStyle.Render(m.trackList.View()), tracker)

	return lipgloss.JoinHorizontal(lipgloss.Bottom, sidePanel, tracker)
}
