package tracklist

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/dece2183/yamusic-tui/config"
)

type helpKeyMap struct {
	CursorUp           key.Binding
	CursorDown         key.Binding
	PageUp             key.Binding
	PageDown           key.Binding
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

func newHelpMap() *helpKeyMap {
	controls := config.Current.Controls
	return &helpKeyMap{
		CursorUp:           key.NewBinding(controls.CursorUp.Binding(), controls.CursorUp.Help("up")),
		CursorDown:         key.NewBinding(controls.CursorDown.Binding(), controls.CursorDown.Help("down")),
		PageUp:             key.NewBinding(controls.TracksNextPage.Binding(), controls.TracksNextPage.Help("page up")),
		PageDown:           key.NewBinding(controls.TracksPrevPage.Binding(), controls.TracksPrevPage.Help("page down")),
		Play:               key.NewBinding(controls.Apply.Binding(), controls.Apply.Help("play")),
		LikeUnlike:         key.NewBinding(controls.TracksLike.Binding(), controls.TracksLike.Help("like/unlike")),
		AddToPlaylist:      key.NewBinding(controls.TracksAddToPlaylist.Binding(), controls.TracksAddToPlaylist.Help("add to")),
		RemoveFromPlaylist: key.NewBinding(controls.TracksRemoveFromPlaylist.Binding(), controls.TracksRemoveFromPlaylist.Help("remove")),
		Search:             key.NewBinding(controls.TracksSearch.Binding(), controls.TracksSearch.Help("search")),
		Share:              key.NewBinding(controls.TracksShare.Binding(), controls.TracksShare.Help("share")),
		Shuffle:            key.NewBinding(controls.TracksShuffle.Binding(), controls.TracksShuffle.Help("shuffle")),
		Reload:             key.NewBinding(controls.Reload.Binding(), controls.Reload.Help("reload")),
		HideTracklist:      key.NewBinding(controls.TracksHide.Binding(), controls.TracksHide.Help("hide")),
		ShowHelp:           key.NewBinding(controls.ShowAllKeys.Binding(), controls.ShowAllKeys.Help("show keys")),
		CloseHelp:          key.NewBinding(controls.ShowAllKeys.Binding(), controls.ShowAllKeys.Help("hide keys")),
	}
}

func (k helpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.CursorUp, k.CursorDown, k.Play, k.LikeUnlike, k.ShowHelp}
}

func (k helpKeyMap) FullHelp() [][]key.Binding {
	bindings := [][]key.Binding{
		{k.CursorUp, k.CursorDown, k.PageUp, k.PageDown},
		{k.Play, k.LikeUnlike, k.AddToPlaylist, k.RemoveFromPlaylist},
		{k.Search, k.Share},
	}

	if k.Shafflable {
		bindings[2] = append(bindings[2], k.Shuffle)
	}

	return append(bindings, []key.Binding{k.Reload, k.HideTracklist, k.CloseHelp})
}

func (k helpKeyMap) HelpHeight() int {
	maxLines := 0
	keys := k.FullHelp()
	for i := range keys {
		if len(keys[i]) > maxLines {
			maxLines = len(keys[i])
		}
	}
	return maxLines
}
