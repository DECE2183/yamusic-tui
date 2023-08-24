package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type playlistListItem struct {
	name    string
	kind    uint64
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

func (d playlistListItemDelegate) Update(message tea.Msg, m *list.Model) tea.Cmd {
	item, ok := m.SelectedItem().(playlistListItem)
	if !ok {
		return nil
	}

	msg, ok := message.(tea.KeyMsg)
	if !ok {
		return nil
	}

	if (key.Matches(msg, m.KeyMap.CursorUp) || key.Matches(msg, m.KeyMap.CursorDown)) && item.active {
		go programm.Send(item)
	}

	return nil
}

func (d playlistListItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(playlistListItem)
	if !ok {
		return
	}

	name := item.name
	if len(name) > 27 {
		name = name[:24] + "..."
	}

	if item.active && !item.subitem {
		if index == m.Index() {
			fmt.Fprint(w, sideBoxSelItemStyle.Render(name))
		} else {
			fmt.Fprint(w, sideBoxItemStyle.Render(name))
		}
	} else {
		if item.subitem {
			if index == m.Index() {
				fmt.Fprint(w, sideBoxSelSubItemStyle.Render(name))
			} else {
				fmt.Fprint(w, sideBoxSubItemStyle.Render(name))
			}
		} else {
			if index == m.Index() {
				fmt.Fprint(w, sideBoxSelInactiveItemStyle.Render(name))
			} else {
				fmt.Fprint(w, sideBoxInactiveItemStyle.Render(name))
			}
		}
	}
}
