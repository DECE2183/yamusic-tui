package tracklist

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dece2183/yamusic-tui/ui/style"
)

type ItemDelegate struct {
	likesMap *map[string]bool
	cacheMap *map[string]bool
}

func (d ItemDelegate) Height() int {
	return 3
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
		trackTitle += style.TrackTitleStyle.Strikethrough(true).Render(item.Track.Title)
	}

	trackVersion := style.TrackVersionStyle.Render(" " + item.Track.Version)
	trackTitle = lipgloss.JoinHorizontal(lipgloss.Top, trackTitle, trackVersion)
	trackArtist := style.TrackVersionStyle.Render(item.Artists)

	durTotal := time.Millisecond * time.Duration(item.Track.DurationMs)
	trackTime := style.TrackVersionStyle.Render(fmt.Sprintf("%d:%02d",
		int(durTotal.Minutes()),
		int(durTotal.Seconds())%60,
	))

	var trackLike string
	if (*d.likesMap)[item.Track.Id] {
		trackLike = style.IconLiked
	} else {
		trackLike = style.IconNotLiked
	}

	var trackCache string
	if (*d.cacheMap)[item.Track.Id] {
		trackCache = style.IconCached
	}

	trackAddInfo := style.TrackAddInfoStyle.Render(trackCache + " " + trackLike + " " + trackTime)
	addInfoLen := lipgloss.Width(trackAddInfo)
	maxLen := m.Width() - addInfoLen - 2
	stl := lipgloss.NewStyle().MaxWidth(maxLen - 1)

	trackTitleLen := lipgloss.Width(trackTitle)
	if trackTitleLen > maxLen {
		trackTitle = stl.Render(trackTitle) + "…"
	} else if trackTitleLen < maxLen {
		trackTitle += strings.Repeat(" ", maxLen-trackTitleLen)
	}

	trackArtistLen := lipgloss.Width(trackArtist)
	if trackArtistLen > maxLen {
		trackArtist = stl.Render(trackArtist) + "…"
	} else if trackArtistLen < maxLen {
		trackArtist += strings.Repeat(" ", maxLen-trackArtistLen)
	}

	trackTitle = lipgloss.JoinVertical(lipgloss.Left, trackTitle, trackArtist)
	trackTitle = lipgloss.JoinHorizontal(lipgloss.Top, trackTitle, trackAddInfo)

	if index == m.Index() {
		stl = style.TrackListActiveStyle
	} else {
		stl = style.TrackListStyle
		if index == m.Index()-1 {
			stl = stl.PaddingBottom(0)
		}
		if index%m.Paginator.PerPage == 0 {
			stl = stl.PaddingTop(1)
		}
	}

	fmt.Fprint(w, stl.Render(trackTitle))
}
