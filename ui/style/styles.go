package style

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	PlaylistsSidePanelWidth = 32
	SearchModalWidth        = 56
)

var (
	AccentColor       = lipgloss.Color("#FC0")
	BackgroundColor   = lipgloss.Color("#6b6b6b")
	ActiveTextColor   = lipgloss.Color("#EEE")
	NormalTextColor   = lipgloss.Color("#CCC")
	InactiveTextColor = lipgloss.Color("#888")
)

var (
	IconPlay     = "▶"
	IconStop     = "■"
	IconLiked    = "💛"
	IconNotLiked = "🤍"
)

var (
	AccentTextStyle = lipgloss.NewStyle().Foreground(AccentColor)
)

var (
	DialogTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F4F4F4")).
				MarginBottom(1)
	DialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(AccentColor).
			Padding(1, 2).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)
	DialogHelpStyle = lipgloss.NewStyle().
			PaddingTop(1)
)

var (
	ButtonStyle = lipgloss.NewStyle().
			Foreground(NormalTextColor).
			Background(InactiveTextColor).
			Padding(0, 3).
			MarginTop(1)
	ActiveButtonStyle = ButtonStyle.
				Foreground(InactiveTextColor).
				Background(AccentColor)
)

var (
	SideBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444")).
			Width(32).
			Padding(1, 0)
	SideBoxItemStyle = lipgloss.NewStyle().
				Foreground(NormalTextColor).
				PaddingLeft(2).
				Width(32).
				MaxWidth(32)
	SideBoxSelItemStyle = SideBoxItemStyle.
				Foreground(ActiveTextColor).
				Background(lipgloss.Color("#4a3c00")).
				PaddingLeft(1).
				Border(lipgloss.InnerHalfBlockBorder()).
				BorderForeground(AccentColor).
				BorderTop(false).
				BorderLeft(true).
				BorderRight(false).
				BorderBottom(false)
	SideBoxInactiveItemStyle = SideBoxItemStyle.
					Foreground(InactiveTextColor).
					Padding(0, 0, 0, 2)
	SideBoxSubItemStyle = SideBoxItemStyle.
				Padding(0, 0, 0, 4)
	SideBoxSelSubItemStyle = SideBoxSelItemStyle.
				Padding(0, 0, 0, 3)
)

var (
	TrackBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444")).
			Padding(1, 2)

	TrackTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#dcdcdc")).
			Bold(true)
	TrackVersionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#999999"))
	TrackArtistStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#dcdcdc"))

	TrackProgressStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				PaddingBottom(1)

	TrackAddInfoStyle = lipgloss.NewStyle().
				Align(lipgloss.Right).
				Width(26)
)

var (
	TrackListStyle = lipgloss.NewStyle().
			Padding(1, 2).
			MarginTop(-2)
	TrackListActiveStyle = lipgloss.NewStyle().
				Padding(0, 1).
				MarginTop(-2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(AccentColor)
)
