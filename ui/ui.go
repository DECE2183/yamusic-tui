package ui

import (
	"fmt"
	"io"
	"yamusic/api"
	"yamusic/config"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type page uint

const (
	_PAGE_LOGIN page = iota
	_PAGE_MAIN  page = iota
	_PAGE_QUIT  page = iota
)

type sideListItem struct {
	name    string
	active  bool
	subitem bool
}

type sideListItemDelegate struct{}

type model struct {
	client *api.YaMusicClient

	width, height int
	page          page

	loginTextInput textinput.Model
	sideList       list.Model
}

func Run(client *api.YaMusicClient) {
	m := model{
		client: client,
		page:   _PAGE_MAIN,

		loginTextInput: textinput.New(),
		sideList:       list.New([]list.Item{}, sideListItemDelegate{}, 512, 512),
	}

	m.loginTextInput.Width = 46
	m.loginTextInput.CharLimit = 39

	sideListItems := []list.Item{
		sideListItem{"wave", true, false},
		sideListItem{"likes", true, false},
		sideListItem{"dislikes", true, false},
		sideListItem{"playlists:", false, false},
	}

	if config.GetToken() == "" {
		m.page = _PAGE_LOGIN
		m.loginTextInput.Focus()
	} else {
		playlists, err := m.client.ListPlaylists()
		if err == nil {
			for _, playlist := range playlists {
				sideListItems = append(sideListItems, sideListItem{playlist.Title, true, true})
			}
			m.sideList.SetItems(sideListItems)
		}
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
	case _PAGE_MAIN:
		m.sideList, cmd = m.sideList.Update(message)
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
	m.sideList.SetSize(28, h)
}

func (i sideListItem) FilterValue() string {
	return i.name
}

func (d sideListItemDelegate) Height() int {
	return 1
}
func (d sideListItemDelegate) Spacing() int {
	return 0
}
func (d sideListItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
func (d sideListItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(sideListItem)
	if !ok {
		return
	}

	if item.active && !item.subitem {
		if index == m.Index() {
			fmt.Fprint(w, sideBoxSelItemStyle.Render(item.name))
		} else {
			fmt.Fprint(w, sideBoxItemStyle.Render(item.name))
		}
	} else {
		if item.subitem {
			if index == m.Index() {
				fmt.Fprint(w, sideBoxSelSubItemStyle.Render(item.name))
			} else {
				fmt.Fprint(w, sideBoxSubItemStyle.Render(item.name))
			}
		} else {
			if index == m.Index() {
				fmt.Fprint(w, sideBoxSelInactiveItemStyle.Render(item.name))
			} else {
				fmt.Fprint(w, sideBoxInactiveItemStyle.Render(item.name))
			}
		}
	}
}
