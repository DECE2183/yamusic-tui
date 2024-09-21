package input

import (
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
)

type Model struct {
	input         textinput.Model
	width, height int
	value         string
}

func New() *Model {
	m := &Model{}
	m.input = textinput.New()
	m.input.Focus()
	return m
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		style.DialogBoxStyle.MaxWidth(m.width).Render(m.input.View()),
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

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.input.Width = w - 9
}

func (m *Model) Value() string {
	return m.value
}
