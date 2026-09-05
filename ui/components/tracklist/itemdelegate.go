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

	var content string
	if item.Album != nil {
		albumMeta := fmt.Sprintf("%d tracks", item.Album.TrackCount)
		if item.Album.Year > 0 {
			albumMeta = fmt.Sprintf("%d · %s", item.Album.Year, albumMeta)
		}
		albumMeta = style.TrackVersionStyle.Render(albumMeta)

		addInfoLen := lipgloss.Width(albumMeta)
		maxLen := m.Width() - addInfoLen - 2
		clip := lipgloss.NewStyle().MaxWidth(maxLen - 1)

		var albumTitle string
		if item.IsPlaying {
			albumTitle = style.AccentTextStyle.Render(style.IconPlay) + " "
		}
		albumTitle += style.TrackTitleStyle.Render(item.Album.Title)
		albumTitleLen := lipgloss.Width(albumTitle)
		if albumTitleLen > maxLen {
			albumTitle = clip.Render(albumTitle) + "…"
		} else if albumTitleLen < maxLen {
			albumTitle += strings.Repeat(" ", maxLen-albumTitleLen)
		}

		albumArtist := style.TrackArtistStyle.Render(item.Artists)
		albumArtistLen := lipgloss.Width(albumArtist)
		if albumArtistLen > maxLen {
			albumArtist = clip.Render(albumArtist) + "…"
		} else if albumArtistLen < maxLen {
			albumArtist += strings.Repeat(" ", maxLen-albumArtistLen)
		}

		content = lipgloss.JoinVertical(lipgloss.Left, albumTitle, albumArtist)
		content = lipgloss.JoinHorizontal(lipgloss.Top, content, albumMeta)
	} else {
		var (
			trackTitle       string
			trackTitleStyle  lipgloss.Style
			trackArtistStyle lipgloss.Style
		)

		if item.IsSuggestion {
			trackTitleStyle = style.TrackTitleStyle.Foreground(style.InactiveTextColor)
			trackArtistStyle = style.TrackArtistStyle.Foreground(style.InactiveTextColor)
			trackTitle += "~ "
		} else {
			trackTitleStyle = style.TrackTitleStyle
			trackArtistStyle = style.TrackArtistStyle
		}

		if item.IsPlaying {
			trackTitle = style.AccentTextStyle.Render(style.IconPlay) + " "
		}
		if item.Track.Available {
			trackTitle += trackTitleStyle.Render(item.Track.Title)
		} else {
			trackTitle += trackTitleStyle.Strikethrough(true).Render(item.Track.Title)
		}

		trackVersion := style.TrackVersionStyle.Render(" " + item.Track.Version)
		trackTitle = lipgloss.JoinHorizontal(lipgloss.Top, trackTitle, trackVersion)
		trackArtist := trackArtistStyle.Render(item.Artists)

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
		clip := lipgloss.NewStyle().MaxWidth(maxLen - 1)

		trackTitleLen := lipgloss.Width(trackTitle)
		if trackTitleLen > maxLen {
			trackTitle = clip.Render(trackTitle) + "…"
		} else if trackTitleLen < maxLen {
			trackTitle += strings.Repeat(" ", maxLen-trackTitleLen)
		}

		trackArtistLen := lipgloss.Width(trackArtist)
		if trackArtistLen > maxLen {
			trackArtist = clip.Render(trackArtist) + "…"
		} else if trackArtistLen < maxLen {
			trackArtist += strings.Repeat(" ", maxLen-trackArtistLen)
		}

		content = lipgloss.JoinVertical(lipgloss.Left, trackTitle, trackArtist)
		content = lipgloss.JoinHorizontal(lipgloss.Top, content, trackAddInfo)
	}

	var stl lipgloss.Style
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

	fmt.Fprint(w, stl.Render(content))
}
