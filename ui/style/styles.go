package style

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dece2183/yamusic-tui/config"
)

const (
	PlaylistsSidePanelWidth = 32
	SearchModalWidth        = 56
)

var (
	AccentColor            lipgloss.Color
	ErrorColor             lipgloss.Color
	BackgroundColor        lipgloss.Color
	ActiveTextColor        lipgloss.Color
	NormalTextColor        lipgloss.Color
	InactiveTextColor      lipgloss.Color
	LyricsPreviosTextColor lipgloss.Color
	LyricsCurrentTextColor lipgloss.Color
	LyricsNextTextColor    lipgloss.Color
)

var (
	IconPlay     = "▶"
	IconStop     = "■"
	IconLiked    = "💛"
	IconNotLiked = "🤍"
	IconCached   = "💿"
	IconDotLight = lipgloss.NewStyle().Foreground(LyricsCurrentTextColor).Render("•")
	IconDotDark  = lipgloss.NewStyle().Foreground(LyricsPreviosTextColor).Render("•")
)

var (
	AccentTextStyle lipgloss.Style
	ErrorTextStyle  lipgloss.Style
)

var (
	DialogTitleStyle lipgloss.Style
	DialogBoxStyle   lipgloss.Style
	DialogHelpStyle  lipgloss.Style
)

var (
	ButtonStyle       lipgloss.Style
	ActiveButtonStyle lipgloss.Style
)

var (
	SideBoxStyle             lipgloss.Style
	SideBoxItemStyle         lipgloss.Style
	SideBoxSelItemStyle      lipgloss.Style
	SideBoxInactiveItemStyle lipgloss.Style
	SideBoxSubItemStyle      lipgloss.Style
	SideBoxSelSubItemStyle   lipgloss.Style
)

var (
	TrackBoxStyle      lipgloss.Style
	TrackTitleStyle    lipgloss.Style
	TrackVersionStyle  lipgloss.Style
	TrackArtistStyle   lipgloss.Style
	TrackProgressStyle lipgloss.Style
	TrackAddInfoStyle  lipgloss.Style
)

var (
	TrackListTitleStyle  lipgloss.Style
	TrackListStyle       lipgloss.Style
	TrackListActiveStyle lipgloss.Style
)

func InitStyles() {
	AccentColor = lipgloss.Color(config.Current.Colors.Accent)
	ErrorColor = lipgloss.Color(config.Current.Colors.Error)
	BackgroundColor = lipgloss.Color(config.Current.Colors.Background)
	ActiveTextColor = lipgloss.Color(config.Current.Colors.ActiveText)
	NormalTextColor = lipgloss.Color(config.Current.Colors.NormalText)
	InactiveTextColor = lipgloss.Color(config.Current.Colors.InactiveText)
	LyricsPreviosTextColor = lipgloss.Color(config.Current.Colors.LyricsPrevious)
	LyricsCurrentTextColor = lipgloss.Color(config.Current.Colors.LyricsCurrent)
	LyricsNextTextColor = lipgloss.Color(config.Current.Colors.LyricsNext)

	AccentTextStyle = lipgloss.NewStyle().Foreground(AccentColor)
	ErrorTextStyle = lipgloss.NewStyle().Foreground(ErrorColor)

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
		PaddingLeft(2).
		PaddingTop(1)

	ButtonStyle = lipgloss.NewStyle().
		Foreground(NormalTextColor).
		Background(InactiveTextColor).
		Padding(0, 3).
		MarginTop(1)
	ActiveButtonStyle = ButtonStyle.
		Foreground(InactiveTextColor).
		Background(AccentColor)

	SideBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#444")).
		Width(PlaylistsSidePanelWidth).
		Padding(1, 0)
	SideBoxItemStyle = lipgloss.NewStyle().
		Foreground(NormalTextColor).
		PaddingLeft(2).
		Width(PlaylistsSidePanelWidth).
		MaxWidth(PlaylistsSidePanelWidth)
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
		Align(lipgloss.Right)

	TrackListTitleStyle = lipgloss.NewStyle().
		Foreground(NormalTextColor).
		UnsetBackground()
	TrackListStyle = lipgloss.NewStyle().
		Padding(0, 2, 1)
	TrackListActiveStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(AccentColor)
}
