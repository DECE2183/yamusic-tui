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

var helpMap = helpKeyMap{
	CursorUp:   key.NewBinding(config.Current.Controls.PlaylistsUp.Binding(), config.Current.Controls.PlaylistsUp.Help("up")),
	CursorDown: key.NewBinding(config.Current.Controls.PlaylistsDown.Binding(), config.Current.Controls.PlaylistsDown.Help("down")),
	Rename:     key.NewBinding(config.Current.Controls.PlaylistsRename.Binding(), config.Current.Controls.PlaylistsRename.Help("rename")),
}
