package search

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dece2183/yamusic-tui/ui/style"
)

type ItemDelegate struct{}

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

	if index == m.Index() {
		fmt.Fprint(w, style.TrackListActiveStyle.Copy().Width(m.Width()-2).MaxWidth(m.Width()).Render(string(item)))
	} else {
		fmt.Fprint(w, style.TrackListStyle.MaxWidth(m.Width()-2).Render(string(item)))
	}
}
