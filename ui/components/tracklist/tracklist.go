package tracklist

import (
	"fmt"
	"io"
	"time"

	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/config"
	"github.com/dece2183/yamusic-tui/ui/helpers"
	"github.com/dece2183/yamusic-tui/ui/model"
	"github.com/dece2183/yamusic-tui/ui/style"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TracklistControl uint

const (
	PLAY TracklistControl = iota
	CURSOR_UP
	CURSOR_DOWN
	SHARE
	LIKE
)

type Item struct {
	Track     *api.Track
	Artists   string
	IsPlaying bool
}

func NewItem(track *api.Track) Item {
	return Item{
		Track:   track,
		Artists: helpers.ArtistList(track.Artists),
	}
}

type ItemDelegate struct {
	likesMap *map[string]bool
}

func (i Item) FilterValue() string {
	return i.Track.Title
}

func (d ItemDelegate) Height() int {
	return 4
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

	var trackTitle string
	if item.IsPlaying {
		trackTitle = style.AccentTextStyle.Render(style.IconPlay) + " "
	}
	if item.Track.Available {
		trackTitle += style.TrackTitleStyle.Render(item.Track.Title)
	} else {
		trackTitle += style.TrackTitleStyle.Copy().Strikethrough(true).Render(item.Track.Title)
	}
	trackVersion := style.TrackVersionStyle.Render(" " + item.Track.Version)
	trackArtist := style.TrackVersionStyle.Render(item.Artists)

	durTotal := time.Millisecond * time.Duration(item.Track.DurationMs)
	trackTime := style.TrackVersionStyle.Render(fmt.Sprintf("%d:%02d",
		int(durTotal.Minutes()),
		int(durTotal.Seconds())%60,
	))

	var trackLike string
	if (*d.likesMap)[item.Track.Id] {
		trackLike = style.IconLiked + " "
	} else {
		trackLike = style.IconNotLiked + " "
	}

	trackAddInfo := style.TrackAddInfoStyle.Render(trackLike + trackTime)

	trackTitle = lipgloss.JoinHorizontal(lipgloss.Top, trackTitle, trackVersion)
	trackTitle = lipgloss.JoinVertical(lipgloss.Left, trackTitle, trackArtist)
	trackTitle = lipgloss.NewStyle().Width(m.Width() - 18).Render(trackTitle)
	trackTitle = lipgloss.JoinHorizontal(lipgloss.Top, trackTitle, trackAddInfo)

	if index == m.Index() {
		fmt.Fprint(w, style.TrackListActiveStyle.Render(trackTitle))
	} else {
		fmt.Fprint(w, style.TrackListStyle.Render(trackTitle))
	}
}

type Model struct {
	program       *tea.Program
	list          list.Model
	width, height int
}

func New(p *tea.Program, likesMap *map[string]bool) Model {
	m := Model{
		program: p,
	}

	controls := config.Current.Controls

	m.list = list.New([]list.Item{}, ItemDelegate{likesMap}, 512, 512)
	m.list.Title = "Tracks"
	m.list.Styles.Title = m.list.Styles.Title.Foreground(style.NormalTextColor).UnsetBackground().Padding(0)
	m.list.KeyMap = list.KeyMap{
		Filter:       key.NewBinding(key.WithKeys(""), controls.Apply.Help("select")),
		CursorUp:     key.NewBinding(controls.TrackListUp.Binding(), controls.TrackListUp.Help("up")),
		CursorDown:   key.NewBinding(controls.TrackListDown.Binding(), controls.TrackListDown.Help("down")),
		Quit:         key.NewBinding(key.WithKeys(""), controls.TrackListLike.Help("like/unlike")),
		ShowFullHelp: key.NewBinding(key.WithKeys(""), controls.TrackListShare.Help("share")),
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
		case controls.Apply.Contains(keypress):
			cmds = append(cmds, model.Cmd(PLAY))
		case controls.TrackListUp.Contains(keypress):
			cmds = append(cmds, model.Cmd(CURSOR_UP))
		case controls.TrackListDown.Contains(keypress):
			cmds = append(cmds, model.Cmd(CURSOR_DOWN))
		case controls.TrackListShare.Contains(keypress):
			cmds = append(cmds, model.Cmd(SHARE))
		case controls.TrackListLike.Contains(keypress):
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
