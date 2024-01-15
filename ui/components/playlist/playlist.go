package playlist

import (
	"fmt"
	"io"
	"yamusic/api"
	"yamusic/config"
	"yamusic/ui/model"
	"yamusic/ui/style"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type PlaylistControl uint

const (
	CURSOR_UP PlaylistControl = iota
	CURSOR_DOWN
)

type PlaylistType = uint64

const (
	NONE PlaylistType = iota
	MYWAVE
	LIKES
)

type Item struct {
	Name         string
	Kind         uint64
	StationId    api.StationId
	StationBatch string
	Active       bool
	Subitem      bool
	Infinite     bool

	Tracks        []api.Track
	CurrentTrack  int
	SelectedTrack int
}

type ItemDelegate struct {
	programm *tea.Program
}

func (i Item) FilterValue() string {
	return i.Name
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
	if len(name) > 27 {
		name = name[:24] + "..."
	}

	if item.Active && !item.Subitem {
		if index == m.Index() {
			fmt.Fprint(w, style.SideBoxSelItemStyle.Render(name))
		} else {
			fmt.Fprint(w, style.SideBoxItemStyle.Render(name))
		}
	} else {
		if item.Subitem {
			if index == m.Index() {
				fmt.Fprint(w, style.SideBoxSelSubItemStyle.Render(name))
			} else {
				fmt.Fprint(w, style.SideBoxSubItemStyle.Render(name))
			}
		} else {
			if index == m.Index() {
				fmt.Fprint(w, style.SideBoxSelInactiveItemStyle.Render(name))
			} else {
				fmt.Fprint(w, style.SideBoxInactiveItemStyle.Render(name))
			}
		}
	}
}

type Model struct {
	program       *tea.Program
	list          list.Model
	width, height int
}

func New(p *tea.Program) Model {
	m := Model{
		program: p,
	}

	playlistItems := []list.Item{
		Item{Name: "my wave", Kind: MYWAVE, Active: true, Subitem: false, Infinite: true},
		Item{Name: "likes", Kind: LIKES, Active: true, Subitem: false},
		Item{Name: "playlists:", Kind: NONE, Active: false, Subitem: false},
	}

	controls := config.Current.Controls

	m.list = list.New(playlistItems, ItemDelegate{programm: p}, 512, 512)
	m.list.Title = "Playlists"
	m.list.SetShowStatusBar(false)
	m.list.Styles.Title = m.list.Styles.Title.Foreground(style.AccentColor).UnsetBackground().Padding(0)
	m.list.KeyMap = list.KeyMap{
		CursorUp:   key.NewBinding(controls.PlaylistsUp.Binding(), controls.PlaylistsUp.Help("up")),
		CursorDown: key.NewBinding(controls.PlaylistsDown.Binding(), controls.PlaylistsUp.Help("down")),
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) View() string {
	return m.list.View()
}

func (m Model) Update(message tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := message.(type) {
	case tea.KeyMsg:
		controls := config.Current.Controls
		keypress := msg.String()

		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)

		switch {
		case controls.PlaylistsUp.Contains(keypress):
			cmds = append(cmds, model.Cmd(CURSOR_UP))
		case controls.PlaylistsDown.Contains(keypress):
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
