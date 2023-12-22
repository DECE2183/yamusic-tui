package loginpage

import (
	"yamusic/config"
	"yamusic/ui/style"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	program *tea.Program

	width, height  int
	loginTextInput textinput.Model
}

// loginpage.Model constructor.
func New() *Model {
	m := &Model{
		loginTextInput: textinput.New(),
	}

	m.loginTextInput.Width = 64
	m.loginTextInput.CharLimit = 60

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	m.program = p

	return m
}

//
// modal.Modal interface implementation
//

func (m *Model) Run() error {
	_, err := m.program.Run()
	if err != nil {
		return err
	}

	config.Current.Token = m.loginTextInput.Value()
	err = config.Save()
	if err != nil {
		return err
	}

	return nil
}

func (m *Model) Send(msg tea.Msg) {
	go m.program.Send(msg)
}

//
// tea.Modal interface implementation
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
		keypress := msg.String()
		switch keypress {
		case "esc", "ctrl+q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, tea.Quit
		default:
			m.loginTextInput, cmd = m.loginTextInput.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	title := style.DialogTitleStyle.Render("Enter your token")
	buttons := lipgloss.Place(42, 1, lipgloss.Right, lipgloss.Center, style.ActiveButtonStyle.Render("Ok"))
	content := lipgloss.JoinVertical(lipgloss.Left, title, m.loginTextInput.View(), buttons)

	dialog := lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		style.DialogBoxStyle.Render(content),
	)

	return dialog
}

//
// private methods
//

func (m *Model) resize(width, height int) {
	m.width, m.height = width, height
}
