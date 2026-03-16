package playlist

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/dece2183/yamusic-tui/config"
)

type helpKeyMap struct {
	CursorUp   key.Binding
	CursorDown key.Binding
	Rename     key.Binding
	Renamable  bool
}

func newHelpMap() *helpKeyMap {
	controls := config.Current.Controls
	return &helpKeyMap{
		CursorUp:   key.NewBinding(controls.PlaylistsUp.Binding(), controls.PlaylistsUp.Help("up")),
		CursorDown: key.NewBinding(controls.PlaylistsDown.Binding(), controls.PlaylistsDown.Help("down")),
		Rename:     key.NewBinding(controls.PlaylistsRename.Binding(), controls.PlaylistsRename.Help("rename")),
	}
}

func (k helpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.CursorUp, k.CursorDown}
}

func (k helpKeyMap) FullHelp() [][]key.Binding {
	if k.Renamable {
		return [][]key.Binding{
			k.ShortHelp(),
			{k.Rename},
		}
	} else {
		return [][]key.Binding{
			k.ShortHelp(),
		}
	}
}
