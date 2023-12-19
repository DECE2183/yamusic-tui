package ui

import (
	"fmt"
	"time"
	"yamusic/api"
	"yamusic/config"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

type trackerHelpKeyMap struct {
	PlayPause  key.Binding
	PrevTrack  key.Binding
	NextTrack  key.Binding
	LikeUnlike key.Binding
	Forward    key.Binding
	Backward   key.Binding
}

var trackerHelpMap = trackerHelpKeyMap{
	PlayPause: key.NewBinding(
		config.Current.Controls.PlayerPause.Binding(),
		config.Current.Controls.PlayerPause.Help("play/pause"),
	),
	PrevTrack: key.NewBinding(
		config.Current.Controls.PlayerPrevious.Binding(),
		config.Current.Controls.PlayerPrevious.Help("previous track"),
	),
	NextTrack: key.NewBinding(
		config.Current.Controls.PlayerNext.Binding(),
		config.Current.Controls.PlayerNext.Help("next track"),
	),
	LikeUnlike: key.NewBinding(
		config.Current.Controls.PlayerLike.Binding(),
		config.Current.Controls.PlayerLike.Help("like/unlike"),
	),
	Backward: key.NewBinding(
		config.Current.Controls.PlayerRewindBackward.Binding(),
		config.Current.Controls.PlayerRewindBackward.Help(fmt.Sprintf("-%d sec", int(config.Current.RewindDuration))),
	),
	Forward: key.NewBinding(
		config.Current.Controls.PlayerRewindForward.Binding(),
		config.Current.Controls.PlayerRewindForward.Help(fmt.Sprintf("+%d sec", int(config.Current.RewindDuration))),
	),
}

func (k trackerHelpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.PlayPause, k.NextTrack, k.PrevTrack, k.Forward, k.Backward, k.LikeUnlike}
}

func (k trackerHelpKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.PlayPause, k.NextTrack, k.PrevTrack, k.Forward, k.Backward, k.LikeUnlike},
	}
}

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
		playButton = activeButtonStyle.Padding(0, 1).Margin(0).Render(iconStop)
	} else {
		playButton = activeButtonStyle.Padding(0, 1).Margin(0).Render(iconPlay)
	}

	var trackTitle string
	if currentTrack.Available {
		trackTitle = trackTitleStyle.Render(currentTrack.Title)
	} else {
		trackTitle = trackTitleStyle.Copy().Strikethrough(true).Render(currentTrack.Title)
	}

	trackVersion := trackVersionStyle.Render(" " + currentTrack.Version)
	trackArtist := trackArtistStyle.Render(artistList(currentTrack.Artists))

	durTotal := time.Millisecond * time.Duration(currentTrack.DurationMs)
	durEllapsed := time.Millisecond * time.Duration(float64(currentTrack.DurationMs)*m.trackProgress.Percent())
	trackTime := trackVersionStyle.Render(fmt.Sprintf("%02d:%02d/%02d:%02d",
		int(durEllapsed.Minutes()),
		int(durEllapsed.Seconds())%60,
		int(durTotal.Minutes()),
		int(durTotal.Seconds())%60,
	))

	var trackLike string
	if m.likedTracksMap[currentTrack.Id] {
		trackLike = iconLiked + " "
	} else {
		trackLike = iconNotLiked + " "
	}

	trackAddInfo := trackAddInfoStyle.Render(trackLike + trackTime)

	trackTitle = lipgloss.JoinHorizontal(lipgloss.Top, trackTitle, trackVersion)
	trackTitle = lipgloss.JoinVertical(lipgloss.Left, trackTitle, trackArtist, "")
	trackTitle = lipgloss.NewStyle().Width(m.width - m.playlistList.Width() - 34).Render(trackTitle)
	trackTitle = lipgloss.JoinHorizontal(lipgloss.Top, trackTitle, trackAddInfo)

	tracker = trackProgressStyle.Render(m.trackProgress.View())
	tracker = lipgloss.JoinHorizontal(lipgloss.Top, playButton, tracker)
	tracker = lipgloss.JoinVertical(lipgloss.Left, tracker, trackTitle, m.trackerHelp.View(trackerHelpMap))

	tracker = trackBoxStyle.Width(m.width - m.playlistList.Width() - 4).Render(tracker)
	tracker = lipgloss.JoinVertical(lipgloss.Left, trackBoxStyle.Render(m.trackList.View()), tracker)

	return lipgloss.JoinHorizontal(lipgloss.Bottom, sidePanel, tracker)
}
