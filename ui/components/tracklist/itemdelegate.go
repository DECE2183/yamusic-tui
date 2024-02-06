package tracklist

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dece2183/yamusic-tui/ui/style"
)

type ItemDelegate struct {
	likesMap *map[string]bool
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
	if item.IsPlaying {
		trackTitle = style.AccentTextStyle.Render(style.IconPlay) + " "
	}
	if item.Track.Available {
		trackTitle += style.TrackTitleStyle.Render(item.Track.Title)
	} else {
		trackTitle += style.TrackTitleStyle.Copy().Strikethrough(true).Render(item.Track.Title)
	}
	trackVersion := style.TrackVersionStyle.Render(" " + item.Track.Version)
	trackArtist := style.TrackVersionStyle.Render(item.Artists)

	durTotal := time.Millisecond * time.Duration(item.Track.DurationMs)
	trackTime := style.TrackVersionStyle.Render(fmt.Sprintf("%d:%02d",
		int(durTotal.Minutes()),
		int(durTotal.Seconds())%60,
	))

	var trackLike string
	if (*d.likesMap)[item.Track.Id] {
		trackLike = style.IconLiked + " "
	} else {
		trackLike = style.IconNotLiked + " "
	}

	trackAddInfo := style.TrackAddInfoStyle.Render(trackLike + trackTime)

	trackTitle = lipgloss.JoinHorizontal(lipgloss.Top, trackTitle, trackVersion)
	trackTitle = lipgloss.JoinVertical(lipgloss.Left, trackTitle, trackArtist)
	trackTitle = lipgloss.NewStyle().Width(m.Width() - lipgloss.Width(trackAddInfo) - 8).Render(trackTitle)
	trackTitle = lipgloss.JoinHorizontal(lipgloss.Top, trackTitle, trackAddInfo)

	if index == m.Index() {
		fmt.Fprint(w, style.TrackListActiveStyle.MaxWidth(m.Width()).Render(trackTitle))
	} else {
		fmt.Fprint(w, style.TrackListStyle.MaxWidth(m.Width()).Render(trackTitle))
	}
}
