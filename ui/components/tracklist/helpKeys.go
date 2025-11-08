package tracklist

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/dece2183/yamusic-tui/config"
)

type helpKeyMap struct {
	CursorUp           key.Binding
	CursorDown         key.Binding
	Play               key.Binding
	LikeUnlike         key.Binding
	AddToPlaylist      key.Binding
	RemoveFromPlaylist key.Binding
	Search             key.Binding
	Share              key.Binding
	Shuffle            key.Binding
	Reload             key.Binding
	ShowHelp           key.Binding
	CloseHelp          key.Binding
	HideTracklist      key.Binding

	Shafflable bool
}

func (k helpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.CursorUp, k.CursorDown, k.Play, k.LikeUnlike, k.ShowHelp}
}

func (k helpKeyMap) FullHelp() [][]key.Binding {
	bindings := [][]key.Binding{
		{k.CursorUp, k.CursorDown, k.Play},
		{k.LikeUnlike, k.AddToPlaylist, k.RemoveFromPlaylist},
		{k.Search, k.Share},
	}

	if k.Shafflable {
		bindings[2] = append(bindings[2], k.Shuffle)
	}

	bindings = append(bindings, []key.Binding{k.Reload, k.HideTracklist, k.CloseHelp})

	return bindings
}

var helpMap = helpKeyMap{
	CursorUp:           key.NewBinding(config.Current.Controls.CursorUp.Binding(), config.Current.Controls.CursorUp.Help("up")),
	CursorDown:         key.NewBinding(config.Current.Controls.CursorDown.Binding(), config.Current.Controls.CursorDown.Help("down")),
	Play:               key.NewBinding(config.Current.Controls.Apply.Binding(), config.Current.Controls.Apply.Help("play")),
	LikeUnlike:         key.NewBinding(config.Current.Controls.TracksLike.Binding(), config.Current.Controls.TracksLike.Help("like/unlike")),
	AddToPlaylist:      key.NewBinding(config.Current.Controls.TracksAddToPlaylist.Binding(), config.Current.Controls.TracksAddToPlaylist.Help("add to")),
	RemoveFromPlaylist: key.NewBinding(config.Current.Controls.TracksRemoveFromPlaylist.Binding(), config.Current.Controls.TracksRemoveFromPlaylist.Help("remove")),
	Search:             key.NewBinding(config.Current.Controls.TracksSearch.Binding(), config.Current.Controls.TracksSearch.Help("search")),
	Share:              key.NewBinding(config.Current.Controls.TracksShare.Binding(), config.Current.Controls.TracksShare.Help("share")),
	Shuffle:            key.NewBinding(config.Current.Controls.TracksShuffle.Binding(), config.Current.Controls.TracksShuffle.Help("shuffle")),
	Reload:             key.NewBinding(config.Current.Controls.Reload.Binding(), config.Current.Controls.Reload.Help("reload")),
	ShowHelp:           key.NewBinding(config.Current.Controls.ShowAllKeys.Binding(), config.Current.Controls.ShowAllKeys.Help("show keys")),
	CloseHelp:          key.NewBinding(config.Current.Controls.ShowAllKeys.Binding(), config.Current.Controls.ShowAllKeys.Help("hide")),
	HideTracklist:      key.NewBinding(config.Current.Controls.TracksHide.Binding(), config.Current.Controls.TracksHide.Help("hide tracklist")),
}
