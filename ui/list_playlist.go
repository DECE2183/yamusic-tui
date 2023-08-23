package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type playlistListItem struct {
	name    string
	id      uint64
	active  bool
	subitem bool
}

type playlistListItemDelegate struct{}

func (i playlistListItem) FilterValue() string {
	return i.name
}

func (d playlistListItemDelegate) Height() int {
	return 1
}

func (d playlistListItemDelegate) Spacing() int {
	return 0
}

func (d playlistListItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d playlistListItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(playlistListItem)
	if !ok {
		return
	}

	if item.active && !item.subitem {
		if index == m.Index() {
			fmt.Fprint(w, sideBoxSelItemStyle.Render(item.name))
		} else {
			fmt.Fprint(w, sideBoxItemStyle.Render(item.name))
		}
	} else {
		if item.subitem {
			if index == m.Index() {
				fmt.Fprint(w, sideBoxSelSubItemStyle.Render(item.name))
			} else {
				fmt.Fprint(w, sideBoxSubItemStyle.Render(item.name))
			}
		} else {
			if index == m.Index() {
				fmt.Fprint(w, sideBoxSelInactiveItemStyle.Render(item.name))
			} else {
				fmt.Fprint(w, sideBoxInactiveItemStyle.Render(item.name))
			}
		}
	}
}
