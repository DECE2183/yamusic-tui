package tracklist

import (
	"github.com/dece2183/yamusic-tui/config"
	"github.com/dece2183/yamusic-tui/ui/model"
	"github.com/dece2183/yamusic-tui/ui/style"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	ADD_TO_PLAYLIST
)

type helpKeyMap struct {
	CursorUp      key.Binding
	CursorDown    key.Binding
	Play          key.Binding
	LikeUnlike    key.Binding
	AddToPlaylist key.Binding
	Search        key.Binding
	Share         key.Binding
	Shuffle       key.Binding
	ShowHelp      key.Binding
	CloseHelp     key.Binding

	Shafflable bool
}

func (k helpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.CursorUp, k.CursorDown, k.Play, k.LikeUnlike, k.ShowHelp}
}

func (k helpKeyMap) FullHelp() [][]key.Binding {
	if k.Shafflable {
		return [][]key.Binding{
			{k.CursorUp, k.CursorDown},
			{k.Play, k.LikeUnlike, k.AddToPlaylist},
			{k.Search, k.Share, k.Shuffle},
			{k.CloseHelp},
		}
	} else {
		return [][]key.Binding{
			{k.CursorUp, k.CursorDown},
			{k.Play, k.LikeUnlike, k.AddToPlaylist},
			{k.Search, k.Share},
			{k.CloseHelp},
		}
	}
}

var helpMap = helpKeyMap{
	CursorUp:      key.NewBinding(config.Current.Controls.CursorUp.Binding(), config.Current.Controls.CursorUp.Help("up")),
	CursorDown:    key.NewBinding(config.Current.Controls.CursorDown.Binding(), config.Current.Controls.CursorDown.Help("down")),
	Play:          key.NewBinding(config.Current.Controls.Apply.Binding(), config.Current.Controls.Apply.Help("play")),
	LikeUnlike:    key.NewBinding(config.Current.Controls.TracksLike.Binding(), config.Current.Controls.TracksLike.Help("like/unlike")),
	AddToPlaylist: key.NewBinding(config.Current.Controls.TracksAddToPlaylist.Binding(), config.Current.Controls.TracksAddToPlaylist.Help("to playlist")),
	Search:        key.NewBinding(config.Current.Controls.TracksSearch.Binding(), config.Current.Controls.TracksSearch.Help("search")),
	Share:         key.NewBinding(config.Current.Controls.TracksShare.Binding(), config.Current.Controls.TracksShare.Help("share")),
	Shuffle:       key.NewBinding(config.Current.Controls.TracksShuffle.Binding(), config.Current.Controls.TracksShuffle.Help("shuffle")),
	ShowHelp:      key.NewBinding(config.Current.Controls.ShowAllKeys.Binding(), config.Current.Controls.ShowAllKeys.Help("show keys")),
	CloseHelp:     key.NewBinding(config.Current.Controls.ShowAllKeys.Binding(), config.Current.Controls.ShowAllKeys.Help("hide")),
}

type Model struct {
	program       *tea.Program
	list          list.Model
	help          help.Model
	width, height int

	Shufflable bool
}

func New(p *tea.Program, likesMap *map[string]bool) *Model {
	m := &Model{
		program: p,
		help:    help.New(),
	}

	m.list = list.New([]list.Item{}, ItemDelegate{likesMap: likesMap}, 512, 512)
	m.list.Title = "Tracks"
	m.list.Styles.Title = m.list.Styles.Title.Foreground(style.NormalTextColor).UnsetBackground().Padding(0)
	m.list.SetShowHelp(false)

	return m
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) View() string {
	helpMap.Shafflable = m.Shufflable
	if m.help.ShowAll {
		m.list.SetHeight(m.height - 4)
	} else {
		m.list.SetHeight(m.height - 2)
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.list.View(), "", m.help.View(helpMap))
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
		case controls.ShowAllKeys.Contains(keypress):
			m.help.ShowAll = !m.help.ShowAll
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
		case controls.TracksAddToPlaylist.Contains(keypress):
			cmds = append(cmds, model.Cmd(ADD_TO_PLAYLIST))
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
	m.list.SetSize(m.width, m.height-3)
}

func (m *Model) SetWidth(w int) {
	m.width = w
	m.list.SetWidth(m.width)
}

func (m *Model) Width() int {
	return m.width
}

func (m *Model) SetHeight(h int) {
	m.height = h
	m.list.SetHeight(m.height - 3)
}

func (m *Model) Height() int {
	return m.height
}
