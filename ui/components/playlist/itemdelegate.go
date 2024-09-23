package playlist

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dece2183/yamusic-tui/ui/style"
)

type ItemDelegate struct {
	programm *tea.Program
}

func (d ItemDelegate) Height() int {
	return 1
}

func (d ItemDelegate) Spacing() int {
	return 0
}

func (d ItemDelegate) Update(message tea.Msg, m *list.Model) tea.Cmd {
	item, ok := m.SelectedItem().(Item)
	if !ok {
		return nil
	}

	msg, ok := message.(tea.KeyMsg)
	if !ok {
		return nil
	}

	if (key.Matches(msg, m.KeyMap.CursorUp) || key.Matches(msg, m.KeyMap.CursorDown)) && item.Active {
		go d.programm.Send(item)
	}

	return nil
}

func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(Item)
	if !ok {
		return
	}

	name := item.Name
	nameLen, _ := lipgloss.Size(name)
	maxLen := m.Width() - 5
	if nameLen > maxLen {
		name = lipgloss.NewStyle().MaxWidth(maxLen-1).Render(name) + "…"
	}

	if !item.Active {
		if item.Subitem {
			fmt.Fprint(w, style.SideBoxSubItemStyle.Render(name))
		} else {
			fmt.Fprint(w, style.SideBoxInactiveItemStyle.Render(name))
		}
		return
	}

	if item.Subitem {
		if index == m.Index() {
			fmt.Fprint(w, style.SideBoxSelSubItemStyle.Render(name))
		} else {
			fmt.Fprint(w, style.SideBoxSubItemStyle.Render(name))
		}
	} else {
		if index == m.Index() {
			fmt.Fprint(w, style.SideBoxSelItemStyle.Render(name))
		} else {
			fmt.Fprint(w, style.SideBoxItemStyle.Render(name))
		}
	}
}
