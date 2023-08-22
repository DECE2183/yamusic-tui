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
				Background(lipgloss.Color("#FC0"))

	sideBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FC0")).
			Padding(1, 0).
			BorderTop(false).
			BorderLeft(false).
			BorderRight(true).
			BorderBottom(false)
	sideBoxItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#CCC")).
				Padding(0, 2)
	sideBoxSelItemStyle = sideBoxItemStyle.Copy().
				Foreground(lipgloss.Color("#EEE")).
				Padding(0, 1).
				Border(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color("#FC0")).
				BorderTop(false).
				BorderLeft(true).
				BorderRight(false).
				BorderBottom(false)
	sideBoxInactiveItemStyle = sideBoxItemStyle.Copy().
					Foreground(lipgloss.Color("#888")).
					Padding(0, 2, 0, 2).
					MarginTop(1)
	sideBoxSelInactiveItemStyle = sideBoxSelItemStyle.Copy().
					BorderForeground(lipgloss.Color("#888")).
					Foreground(lipgloss.Color("#888")).
					Padding(0, 2, 0, 1).
					MarginTop(1)
	sideBoxSubItemStyle = sideBoxItemStyle.Copy().
				Padding(0, 2, 0, 4)
	sideBoxSelSubItemStyle = sideBoxSelItemStyle.Copy().
				Padding(0, 2, 0, 3)
)
