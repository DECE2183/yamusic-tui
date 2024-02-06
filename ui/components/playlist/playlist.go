package playlist

import (
	"github.com/dece2183/yamusic-tui/config"
	"github.com/dece2183/yamusic-tui/ui/model"
	"github.com/dece2183/yamusic-tui/ui/style"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type Control uint

const (
	CURSOR_UP Control = iota
	CURSOR_DOWN
)

type PlaylistType = uint64

const (
	NONE PlaylistType = iota
	MYWAVE
	LIKES
)

type Model struct {
	program       *tea.Program
	list          list.Model
	width, height int
}

func New(p *tea.Program, title string) *Model {
	m := &Model{
		program: p,
	}

	playlistItems := []list.Item{
		Item{Name: "my wave", Kind: MYWAVE, Active: true, Subitem: false, Infinite: true},
		Item{Name: "likes", Kind: LIKES, Active: true, Subitem: false},

		Item{Name: "", Kind: NONE, Active: false, Subitem: false},
		Item{Name: "playlists:", Kind: NONE, Active: false, Subitem: false},
	}

	controls := config.Current.Controls

	m.list = list.New(playlistItems, ItemDelegate{programm: p}, 512, 512)
	m.list.Title = title
	m.list.SetShowStatusBar(false)
	m.list.Styles.Title = m.list.Styles.Title.Foreground(style.AccentColor).UnsetBackground().Padding(0)
	m.list.KeyMap = list.KeyMap{
		CursorUp:   key.NewBinding(controls.PlaylistsUp.Binding(), controls.PlaylistsUp.Help("up")),
		CursorDown: key.NewBinding(controls.PlaylistsDown.Binding(), controls.PlaylistsDown.Help("down")),
	}

	return m
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) View() string {
	return m.list.View()
}

func (m *Model) Update(message tea.Msg) (*Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := message.(type) {
	case tea.KeyMsg:
		controls := config.Current.Controls
		keypress := msg.String()

		switch {
		case controls.PlaylistsUp.Contains(keypress):
			m.list, cmd = m.list.Update(msg)

			for len(m.list.Items()) > 0 && m.list.Index() > 0 && !m.list.SelectedItem().(Item).Active {
				m.list.CursorUp()
			}

			cmds = append(cmds, cmd)
			cmds = append(cmds, model.Cmd(CURSOR_UP))
		case controls.PlaylistsDown.Contains(keypress):
			m.list, cmd = m.list.Update(msg)

			for m.list.Index() < len(m.list.Items())-1 && !m.list.SelectedItem().(Item).Active {
				m.list.CursorDown()
			}

			cmds = append(cmds, cmd)
			cmds = append(cmds, model.Cmd(CURSOR_DOWN))
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) Items() []Item {
	litems := m.list.Items()
	items := make([]Item, len(litems))
	for i := range litems {
		items[i] = litems[i].(Item)
	}
	return items
}

func (m *Model) SetItems(items []Item) tea.Cmd {
	newItems := make([]list.Item, len(items))
	for i := 0; i < len(items); i++ {
		newItems[i] = items[i]
	}
	return m.list.SetItems(newItems)
}

func (m *Model) InsertItem(index int, item Item) tea.Cmd {
	if index < 0 {
		index = len(m.list.Items()) + 1
	}
	return m.list.InsertItem(index, item)
}

func (m *Model) SetItem(index int, item Item) tea.Cmd {
	return m.list.SetItem(index, item)
}

func (m *Model) SelectedItem() Item {
	return m.list.SelectedItem().(Item)
}

func (m *Model) Index() int {
	return m.list.Index()
}

func (m *Model) Select(index int) {
	m.list.Select(index)
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.list.SetSize(m.width, m.height)
}

func (m *Model) SetWidth(w int) {
	m.width = w
	m.list.SetSize(m.width, m.height)
}

func (m *Model) Width() int {
	return m.width
}

func (m *Model) SetHeight(h int) {
	m.height = h
	m.list.SetSize(m.width, m.height)
}

func (m *Model) Height() int {
	return m.height
}
