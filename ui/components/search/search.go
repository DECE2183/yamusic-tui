package search

import (
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
)

var additionalKeyBindigs = []key.Binding{
	key.NewBinding(config.Current.Controls.Apply.Binding(), config.Current.Controls.Apply.Help("search")),
	key.NewBinding(config.Current.Controls.Cancel.Binding(), config.Current.Controls.Cancel.Help("cancel")),
}

type Model struct {
	list          list.Model
	input         textinput.Model
	width, height int
}

func New() *Model {
	m := &Model{}

	controls := config.Current.Controls

	m.list = list.New([]list.Item{}, ItemDelegate{}, 512, 512)
	m.list.Title = ""
	m.list.Styles.Title = lipgloss.NewStyle().Height(0)
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
	return additionalKeyBindigs
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		style.DialogBoxStyle.MaxWidth(m.width).Render(m.input.View()),
		m.list.View(),
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
		case controls.Cancel.Contains(keypress):
			cmds = append(cmds, model.Cmd(CANCEL))
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
		}

	default:
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
		cmds = append(cmds, model.Cmd(TYPING))
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.input.Width = w - 9
	m.list.SetSize(m.width, m.height-3)
}

func (m *Model) SetSuggestions(best string, suggestions []string) {
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
