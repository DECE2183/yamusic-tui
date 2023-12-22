package tracklist

import (
	"fmt"
	"io"
	"time"
	"yamusic/ui/style"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Item struct {
	Title      string
	Version    string
	Artists    string
	Id         string
	DurationMs int
	Liked      bool
	Available  bool
}

type ItemDelegate struct{}

func (i Item) FilterValue() string {
	return i.Title
}

func (d ItemDelegate) Height() int {
	return 4
}

func (d ItemDelegate) Spacing() int {
	return 0
}

func (d ItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(Item)
	if !ok {
		return
	}

	var trackTitle string
	if item.Available {
		trackTitle = style.TrackTitleStyle.Render(item.Title)
	} else {
		trackTitle = style.TrackTitleStyle.Copy().Strikethrough(true).Render(item.Title)
	}
	trackVersion := style.TrackVersionStyle.Render(" " + item.Version)
	trackArtist := style.TrackVersionStyle.Render(item.Artists)

	durTotal := time.Millisecond * time.Duration(item.DurationMs)
	trackTime := style.TrackVersionStyle.Render(fmt.Sprintf("%d:%02d",
		int(durTotal.Minutes()),
		int(durTotal.Seconds())%60,
	))

	var trackLike string
	if item.Liked {
		trackLike = style.IconLiked + " "
	} else {
		trackLike = style.IconNotLiked + " "
	}

	trackAddInfo := style.TrackAddInfoStyle.Render(trackLike + trackTime)

	trackTitle = lipgloss.JoinHorizontal(lipgloss.Top, trackTitle, trackVersion)
	trackTitle = lipgloss.JoinVertical(lipgloss.Left, trackTitle, trackArtist)
	trackTitle = lipgloss.NewStyle().Width(m.Width() - 18).Render(trackTitle)
	trackTitle = lipgloss.JoinHorizontal(lipgloss.Top, trackTitle, trackAddInfo)

	if index == m.Index() {
		fmt.Fprint(w, style.TrackListActiveStyle.Render(trackTitle))
	} else {
		fmt.Fprint(w, style.TrackListStyle.Render(trackTitle))
	}
}
