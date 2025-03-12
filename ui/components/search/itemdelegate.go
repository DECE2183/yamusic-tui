package search

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dece2183/yamusic-tui/ui/style"
)

type ItemDelegate struct{}

func (d ItemDelegate) Height() int {
	return 2
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

	maxWidth := m.Width() - 3
	if index == m.Index() {
		maxWidth -= 1
	}

	text := string(item)
	textLen := lipgloss.Width(text)
	if textLen > maxWidth {
		text = text[:maxWidth-1] + "…"
	} else if textLen < maxWidth {
		text += strings.Repeat(" ", maxWidth-textLen)
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

	fmt.Fprint(w, stl.Render(text))
}
