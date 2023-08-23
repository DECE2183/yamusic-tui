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
)

var (
	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#181818")).
			Background(lipgloss.Color("#FC0")).
			Padding(0, 3).
			MarginTop(1)
	activeButtonStyle = buttonStyle.Copy().
				Foreground(lipgloss.Color("#181818")).
				Background(lipgloss.Color("#FC0"))
)

var (
	sideBoxStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#121212")).
			Width(32).
			Padding(1, 0)
	sideBoxItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#CCC")).
				Background(lipgloss.Color("#121212")).
				PaddingLeft(2)
	sideBoxSelItemStyle = sideBoxItemStyle.Copy().
				Foreground(lipgloss.Color("#EEE")).
				PaddingLeft(1).
				Border(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color("#FC0")).
				BorderTop(false).
				BorderLeft(true).
				BorderRight(false).
				BorderBottom(false)
	sideBoxInactiveItemStyle = sideBoxItemStyle.Copy().
					Foreground(lipgloss.Color("#888")).
					Padding(0, 0, 0, 2).
					MarginTop(1)
	sideBoxSelInactiveItemStyle = sideBoxSelItemStyle.Copy().
					BorderForeground(lipgloss.Color("#888")).
					Foreground(lipgloss.Color("#888")).
					Padding(0, 0, 0, 1).
					MarginTop(1)
	sideBoxSubItemStyle = sideBoxItemStyle.Copy().
				Padding(0, 0, 0, 4)
	sideBoxSelSubItemStyle = sideBoxSelItemStyle.Copy().
				Padding(0, 0, 0, 3)
)

var (
	trackBoxStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#181818")).
			PaddingBottom(1)

	trackTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#dcdcdc")).
			PaddingLeft(2).
			Bold(true)
	trackVersionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#999999")).
				PaddingLeft(2)
	trackArtistStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#dcdcdc")).
				PaddingLeft(2)

	trackProgressStyle = lipgloss.NewStyle().
				PaddingBottom(1)
)
