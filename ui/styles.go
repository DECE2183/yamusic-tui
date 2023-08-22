package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	dialogTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F4F4F4")).
				MarginBottom(1)

	dialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FC0")).
			Padding(1, 2).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true).
			Width(46)

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#181818")).
			Background(lipgloss.Color("#FC0")).
			Padding(0, 3).
			MarginTop(1)

	activeButtonStyle = buttonStyle.Copy().
				Foreground(lipgloss.Color("#181818")).
				Background(lipgloss.Color("#FC0")).
				Padding(0, 3).
				MarginTop(1)
)
