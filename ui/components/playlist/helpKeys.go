package playlist

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/dece2183/yamusic-tui/config"
)

type helpKeyMap struct {
	CursorUp      key.Binding
	CursorDown    key.Binding
	Rename        key.Binding
	HidePlaylists key.Binding
	Renamable     bool
}

func (k helpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.CursorUp, k.CursorDown}
}

func (k helpKeyMap) FullHelp() [][]key.Binding {
	bindings := [][]key.Binding{
		k.ShortHelp(),
	}

	if k.Renamable {
		bindings = append(bindings, []key.Binding{k.Rename})
	}

	bindings = append(bindings, []key.Binding{k.HidePlaylists})

	return bindings
}

var helpMap = helpKeyMap{
	CursorUp:      key.NewBinding(config.Current.Controls.PlaylistsUp.Binding(), config.Current.Controls.PlaylistsUp.Help("up")),
	CursorDown:    key.NewBinding(config.Current.Controls.PlaylistsDown.Binding(), config.Current.Controls.PlaylistsDown.Help("down")),
	Rename:        key.NewBinding(config.Current.Controls.PlaylistsRename.Binding(), config.Current.Controls.PlaylistsRename.Help("rename")),
	HidePlaylists: key.NewBinding(config.Current.Controls.PlaylistsHide.Binding(), config.Current.Controls.PlaylistsHide.Help("hide playlists")),
}
