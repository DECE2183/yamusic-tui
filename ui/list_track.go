package ui

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type trackListItem struct {
	title      string
	version    string
	artists    string
	id         string
	durationMs int
	liked      bool
	available  bool
}

type trackListItemDelegate struct{}

func (i trackListItem) FilterValue() string {
	return i.title
}

func (d trackListItemDelegate) Height() int {
	return 4
}

func (d trackListItemDelegate) Spacing() int {
	return 0
}

func (d trackListItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d trackListItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(trackListItem)
	if !ok {
		return
	}

	var trackTitle string
	if item.available {
		trackTitle = trackTitleStyle.Render(item.title)
	} else {
		trackTitle = trackTitleStyle.Copy().Strikethrough(true).Render(item.title)
	}
	trackVersion := trackVersionStyle.Render(" " + item.version)
	trackArtist := trackVersionStyle.Render(item.artists)

	durTotal := time.Millisecond * time.Duration(item.durationMs)
	trackTime := trackVersionStyle.Render(fmt.Sprintf("%d:%02d",
		int(durTotal.Minutes()),
		int(durTotal.Seconds())%60,
	))

	var trackLike string
	if item.liked {
		trackLike = iconLiked + " "
	} else {
		trackLike = iconNotLiked + " "
	}

	trackAddInfo := trackAddInfoStyle.Render(trackLike + trackTime)

	trackTitle = lipgloss.JoinHorizontal(lipgloss.Top, trackTitle, trackVersion)
	trackTitle = lipgloss.JoinVertical(lipgloss.Left, trackTitle, trackArtist)
	trackTitle = lipgloss.NewStyle().Width(m.Width() - 18).Render(trackTitle)
	trackTitle = lipgloss.JoinHorizontal(lipgloss.Top, trackTitle, trackAddInfo)

	if index == m.Index() {
		fmt.Fprint(w, trackListActiveStyle.Render(trackTitle))
	} else {
		fmt.Fprint(w, trackListStyle.Render(trackTitle))
	}
}
