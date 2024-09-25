package input

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dece2183/yamusic-tui/config"
	"github.com/dece2183/yamusic-tui/ui/model"
	"github.com/dece2183/yamusic-tui/ui/style"
)

type Control uint

const (
	APPLY Control = iota
	CANCEL
)

type helpKeyMap struct {
	apply  key.Binding
	cancel key.Binding
}

func (k helpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.apply, k.cancel}
}

func (k helpKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}

var helpMap = helpKeyMap{
	apply: key.NewBinding(
		config.Current.Controls.Apply.Binding(),
		config.Current.Controls.Apply.Help("apply"),
	),
	cancel: key.NewBinding(
		config.Current.Controls.Cancel.Binding(),
		config.Current.Controls.Cancel.Help("cancel"),
	),
}

type Model struct {
	input textinput.Model
	help  help.Model
	width int
	value string

	Title string
}

func New() *Model {
	m := &Model{
		input: textinput.New(),
		help:  help.New(),
	}
	m.input.Focus()
	return m
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) View() string {
	title := style.DialogTitleStyle.Render(m.Title)
	content := lipgloss.JoinVertical(lipgloss.Left, title, m.input.View())
	return lipgloss.JoinVertical(
		lipgloss.Left,
		style.DialogBoxStyle.Render(content),
		style.DialogHelpStyle.Render(m.help.View(helpMap)),
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
			cmds = append(cmds, model.Cmd(APPLY))
			m.value = m.input.Value()
			m.input.Reset()
		case controls.Cancel.Contains(keypress):
			cmds = append(cmds, model.Cmd(CANCEL))
			m.input.Reset()
		default:
			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)
		}

	default:
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) SetWidth(w int) {
	m.width = w
	m.input.Width = w - 9
}

func (m *Model) Value() string {
	return m.value
}

func (m *Model) SetValue(val string) {
	m.input.SetValue(val)
}
