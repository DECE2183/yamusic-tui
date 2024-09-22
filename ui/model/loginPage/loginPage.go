package loginpage

import (
	"github.com/dece2183/yamusic-tui/config"
	"github.com/dece2183/yamusic-tui/ui/style"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	err           error
	program       *tea.Program
	width, height int

	input textinput.Model
	help  help.Model
}

// loginpage.Model constructor.
func New() *Model {
	m := &Model{
		input: textinput.New(),
		help:  help.New(),
	}

	m.input.Width = 64
	m.input.CharLimit = 64
	m.input.Focus()

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	m.program = p

	return m
}

//
// model.Model interface implementation
//

func (m *Model) Run() error {
	_, err := m.program.Run()
	if err != nil {
		return err
	}
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m *Model) Send(msg tea.Msg) {
	go m.program.Send(msg)
}

//
// tea.Model interface implementation
//

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		return m, tea.ClearScreen

	case tea.KeyMsg:
		controls := config.Current.Controls
		keypress := msg.String()

		switch {
		case controls.Quit.Contains(keypress):
			return m, tea.Quit
		case controls.Apply.Contains(keypress):
			config.Current.Token = m.input.Value()
			m.err = config.Save()
			return m, tea.Quit
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

func (m *Model) View() string {
	title := style.DialogTitleStyle.Render("Enter your token")
	content := lipgloss.JoinVertical(lipgloss.Left, title, m.input.View())
	content = style.DialogBoxStyle.Render(content)
	content = lipgloss.JoinVertical(lipgloss.Left, content, style.DialogHelpStyle.Render(m.help.View(helpMap)))

	dialog := lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)

	return dialog
}

//
// private methods
//

func (m *Model) resize(width, height int) {
	m.width, m.height = width, height
}
