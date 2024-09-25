package search

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dece2183/yamusic-tui/config"
	"github.com/dece2183/yamusic-tui/ui/model"
	"github.com/dece2183/yamusic-tui/ui/style"
)

type Control uint

const (
	SELECT Control = iota
	CANCEL
	CURSOR_UP
	CURSOR_DOWN
	TYPING
	UPDATE_SUGGESTIONS
)

const (
	_UPDATE_SUGGESTIONS_PERIOD = time.Millisecond * 4
)

type Model struct {
	list                 list.Model
	input                textinput.Model
	width, height        int
	value                string
	updated              bool
	lastUpdateTime       time.Time
	additionalKeyBindigs []key.Binding

	Title  string
	Action string
}

func New() *Model {
	m := &Model{
		additionalKeyBindigs: []key.Binding{
			key.NewBinding(config.Current.Controls.Apply.Binding(), config.Current.Controls.Apply.Help("search")),
			key.NewBinding(config.Current.Controls.Cancel.Binding(), config.Current.Controls.Cancel.Help("cancel")),
		},
		Title:  "Search",
		Action: "search",
	}

	controls := config.Current.Controls

	m.list = list.New([]list.Item{}, ItemDelegate{}, 512, 512)
	m.list.SetShowTitle(false)
	m.list.DisableQuitKeybindings()
	m.list.KeyMap = list.KeyMap{
		CursorUp:   key.NewBinding(controls.CursorUp.Binding(), controls.CursorUp.Help("up")),
		CursorDown: key.NewBinding(controls.CursorDown.Binding(), controls.CursorDown.Help("down")),
	}
	m.list.AdditionalShortHelpKeys = m.keymap

	m.input = textinput.New()
	m.input.Focus()

	return m
}

func (m *Model) keymap() []key.Binding {
	return m.additionalKeyBindigs
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) View() string {
	m.additionalKeyBindigs[0].SetHelp(m.additionalKeyBindigs[0].Help().Key, m.Action)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		style.AccentTextStyle.MaxWidth(m.width).MarginBottom(1).Render(m.Title),
		style.DialogBoxStyle.MaxWidth(m.width).Render(m.input.View()),
		lipgloss.NewStyle().MaxWidth(m.width).Render(m.list.View()),
	)
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
		case controls.Apply.Contains(keypress):
			cmds = append(cmds, model.Cmd(SELECT))

			if len(m.list.Items()) == 0 {
				m.value = ""
				break
			}

			suggest, ok := m.list.SelectedItem().(Item)
			if !ok {
				m.value = ""
				break
			}

			m.value = string(suggest)
			m.list.SetItems([]list.Item{})
			m.list.Select(0)
			m.input.Reset()
		case controls.Cancel.Contains(keypress):
			cmds = append(cmds, model.Cmd(CANCEL))
			m.list.SetItems([]list.Item{})
			m.list.Select(0)
			m.input.Reset()
			m.value = ""
		case controls.CursorUp.Contains(keypress):
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
			cmds = append(cmds, model.Cmd(CURSOR_UP))
		case controls.CursorDown.Contains(keypress):
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
			cmds = append(cmds, model.Cmd(CURSOR_DOWN))
		default:
			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)
			cmds = append(cmds, model.Cmd(TYPING))
			m.lastUpdateTime = time.Now()
			m.updated = false
		}

	default:
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)

		if !m.updated && time.Now().After(m.lastUpdateTime.Add(_UPDATE_SUGGESTIONS_PERIOD)) {
			cmds = append(cmds, model.Cmd(UPDATE_SUGGESTIONS))
			m.updated = true
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.input.Width = w - 9
	m.list.SetSize(m.width, m.height-6)
}

func (m *Model) SetSuggestions(suggestions []string) {
	items := make([]list.Item, 0, len(suggestions)+1)

	if len(m.input.Value()) > 0 {
		items = append(items, Item(m.input.Value()))
	}

	for _, sug := range suggestions {
		items = append(items, Item(sug))
	}

	m.list.SetItems(items)
}

func (m *Model) InputValue() string {
	return m.input.Value()
}

func (m *Model) SuggestionValue() (string, bool) {
	return m.value, len(m.value) > 0
}
