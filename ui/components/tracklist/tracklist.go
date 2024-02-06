package tracklist

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
	PLAY Control = iota
	CURSOR_UP
	CURSOR_DOWN
	SEARCH
	SHUFFLE
	SHARE
	LIKE
)

var additionalKeyBindigs = []key.Binding{
	key.NewBinding(config.Current.Controls.Apply.Binding(), config.Current.Controls.Apply.Help("play")),
	key.NewBinding(config.Current.Controls.TracksLike.Binding(), config.Current.Controls.TracksLike.Help("like/unlike")),
	key.NewBinding(config.Current.Controls.TracksShare.Binding(), config.Current.Controls.TracksShare.Help("share")),
}

type Model struct {
	program       *tea.Program
	list          list.Model
	width, height int

	Shufflable bool
}

func New(p *tea.Program, likesMap *map[string]bool) *Model {
	m := &Model{
		program: p,
	}

	controls := config.Current.Controls

	m.list = list.New([]list.Item{}, ItemDelegate{likesMap: likesMap}, 512, 512)
	m.list.Title = "Tracks"
	m.list.Styles.Title = m.list.Styles.Title.Foreground(style.NormalTextColor).UnsetBackground().Padding(0)
	m.list.KeyMap = list.KeyMap{
		CursorUp:   key.NewBinding(controls.CursorUp.Binding(), controls.CursorUp.Help("up")),
		CursorDown: key.NewBinding(controls.CursorDown.Binding(), controls.CursorDown.Help("down")),
	}
	m.list.AdditionalShortHelpKeys = m.keymap

	return m
}

func (m *Model) keymap() []key.Binding {
	controls := config.Current.Controls

	if m.Shufflable {
		return append(additionalKeyBindigs, key.NewBinding(controls.TracksShuffle.Binding(), controls.TracksShuffle.Help("shuffle")))
	}

	return additionalKeyBindigs
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

		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)

		switch {
		case controls.Apply.Contains(keypress):
			cmds = append(cmds, model.Cmd(PLAY))
		case controls.CursorUp.Contains(keypress):
			cmds = append(cmds, model.Cmd(CURSOR_UP))
		case controls.CursorDown.Contains(keypress):
			cmds = append(cmds, model.Cmd(CURSOR_DOWN))
		case controls.TracksSearch.Contains(keypress):
			cmds = append(cmds, model.Cmd(SEARCH))
		case controls.TracksShuffle.Contains(keypress):
			cmds = append(cmds, model.Cmd(SHUFFLE))
		case controls.TracksShare.Contains(keypress):
			cmds = append(cmds, model.Cmd(SHARE))
		case controls.TracksLike.Contains(keypress):
			cmds = append(cmds, model.Cmd(LIKE))
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

func (m *Model) RemoveItem(index int) {
	m.list.RemoveItem(index)
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
