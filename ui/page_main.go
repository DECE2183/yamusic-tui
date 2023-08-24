package ui

import (
	"fmt"
	"time"
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
		playButton = activeButtonStyle.Padding(0, 1).Margin(0).Render("■")
	} else {
		playButton = activeButtonStyle.Padding(0, 1).Margin(0).Render("▶")
	}

	trackTitle := trackTitleStyle.Render(currentTrack.Title)
	trackVersion := trackVersionStyle.Render(" " + currentTrack.Version)
	trackArtist := trackArtistStyle.Render(artistList(currentTrack.Artists))

	durTotal := time.Millisecond * time.Duration(currentTrack.DurationMs)
	durEllapsed := time.Millisecond * time.Duration(float64(currentTrack.DurationMs)*m.trackProgress.Percent())
	trackTime := trackVersionStyle.Copy().Width(26).Align(lipgloss.Right).Render(fmt.Sprintf("%02d:%02d/%02d:%02d",
		int(durEllapsed.Minutes()),
		int(durEllapsed.Seconds())%60,
		int(durTotal.Minutes()),
		int(durTotal.Seconds())%60,
	))

	trackTitle = lipgloss.JoinHorizontal(lipgloss.Top, trackTitle, trackVersion)
	trackTitle = lipgloss.JoinVertical(lipgloss.Left, trackTitle, trackArtist)
	trackTitle = lipgloss.NewStyle().Width(m.width - m.playlistList.Width() - 34).Render(trackTitle)
	trackTitle = lipgloss.JoinHorizontal(lipgloss.Top, trackTitle, trackTime)

	tracker = trackProgressStyle.Render(m.trackProgress.View())
	tracker = lipgloss.JoinHorizontal(lipgloss.Top, playButton, tracker)
	tracker = lipgloss.JoinVertical(lipgloss.Left, tracker, trackTitle)

	tracker = trackBoxStyle.Width(m.width - m.playlistList.Width() - 4).Render(tracker)
	tracker = lipgloss.JoinVertical(lipgloss.Left, trackBoxStyle.Render(m.trackList.View()), tracker)

	return lipgloss.JoinHorizontal(lipgloss.Bottom, sidePanel, tracker)
}
