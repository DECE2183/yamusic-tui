package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type trackListItem struct {
	title   string
	version string
	artists string
	id      string
	liked   bool
}

type trackListItemDelegate struct{}

func (i trackListItem) FilterValue() string {
	return i.title
}

func (d trackListItemDelegate) Height() int {
	return 1
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
	fmt.Fprint(w, sideBoxInactiveItemStyle.Render(item.title))
}
