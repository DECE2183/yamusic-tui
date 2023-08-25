package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	accentColor       = lipgloss.Color("#FC0")
	activeTextColor   = lipgloss.Color("#EEE")
	normalTextColor   = lipgloss.Color("#CCC")
	inactiveTextColor = lipgloss.Color("#888")
)

var (
	iconPlay     = "▶"
	iconStop     = "■"
	iconLiked    = "💛"
	iconNotLiked = "🤍"
)

var (
	dialogTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F4F4F4")).
				MarginBottom(1)
	dialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentColor).
			Padding(1, 2).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true).
			Width(46)
)

var (
	buttonStyle = lipgloss.NewStyle().
			Foreground(normalTextColor).
			Background(accentColor).
			Padding(0, 3).
			MarginTop(1)
	activeButtonStyle = buttonStyle.Copy().
				Foreground(normalTextColor).
				Background(accentColor)
)

var (
	sideBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444")).
			Width(32).
			Padding(1, 0)
	sideBoxItemStyle = lipgloss.NewStyle().
				Foreground(normalTextColor).
				PaddingLeft(2).
				Width(32).
				MaxWidth(32)
	sideBoxSelItemStyle = sideBoxItemStyle.Copy().
				Foreground(activeTextColor).
				Background(lipgloss.Color("#4a3c00")).
				PaddingLeft(1).
				Border(lipgloss.InnerHalfBlockBorder()).
				BorderForeground(accentColor).
				BorderTop(false).
				BorderLeft(true).
				BorderRight(false).
				BorderBottom(false)
	sideBoxInactiveItemStyle = sideBoxItemStyle.Copy().
					Foreground(inactiveTextColor).
					Padding(0, 0, 0, 2).
					MarginTop(1)
	sideBoxSelInactiveItemStyle = sideBoxSelItemStyle.Copy().
					BorderForeground(inactiveTextColor).
					UnsetBackground().
					Foreground(inactiveTextColor).
					Padding(0, 0, 0, 1).
					MarginTop(1)
	sideBoxSubItemStyle = sideBoxItemStyle.Copy().
				Padding(0, 0, 0, 4)
	sideBoxSelSubItemStyle = sideBoxSelItemStyle.Copy().
				Padding(0, 0, 0, 3)
)

var (
	trackBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444")).
			Padding(1, 2)

	trackTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#dcdcdc")).
			Bold(true)
	trackVersionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#999999"))
	trackArtistStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#dcdcdc"))

	trackProgressStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				PaddingBottom(1)

	trackAddInfoStyle = lipgloss.NewStyle().
				Align(lipgloss.Right).
				Width(26)
)

var (
	trackListStyle = lipgloss.NewStyle().
			Padding(1, 2).
			MarginTop(-2)
	trackListActiveStyle = lipgloss.NewStyle().
				Padding(0, 1).
				MarginTop(-2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(accentColor)
)
