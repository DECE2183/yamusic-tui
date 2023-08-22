package ui

import (
	"strings"
	"yamusic/config"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type page uint

const (
	_PAGE_LOGIN page = iota
	_PAGE_MAIN  page = iota
	_PAGE_QUIT  page = iota
)

type model struct {
	width, height int
	page          page

	loginTextInput textinput.Model
}

var doc = strings.Builder{}

func Run() {
	m := model{
		page:           _PAGE_MAIN,
		loginTextInput: textinput.New(),
	}

	m.loginTextInput.Width = 46
	m.loginTextInput.CharLimit = 39

	if config.GetToken() == "" {
		m.page = _PAGE_LOGIN
		m.loginTextInput.Focus()
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	p.Run()
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd tea.Cmd
	)

	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		keypress := msg.String()
		if keypress == "esc" || keypress == "ctrl+q" || keypress == "ctrl+c" {
			m.page = _PAGE_QUIT
			return m, tea.Quit
		}

		switch m.page {
		case _PAGE_LOGIN:
			if keypress == "enter" {
				err := config.SaveToken(m.loginTextInput.Value())
				if err != nil {
					return m, nil
				}
				m.page = _PAGE_MAIN
				return m, nil
			}
		case _PAGE_MAIN:
		}
	}

	switch m.page {
	case _PAGE_LOGIN:
		m.loginTextInput, cmd = m.loginTextInput.Update(message)
	}

	return m, cmd
}

func (m model) View() string {
	switch m.page {
	case _PAGE_LOGIN:
		return m.renderLoginPage()
	case _PAGE_MAIN:
		return m.renderMainPage()
	}

	return ""
}

func (m *model) resize(w, h int) {
	m.width, m.height = w, h
}
