package tracker

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/dece2183/yamusic-tui/config"
)

type helpKeyMap struct {
	PlayPause    key.Binding
	PrevTrack    key.Binding
	NextTrack    key.Binding
	LikeUnlike   key.Binding
	CacheTrack   key.Binding
	Forward      key.Binding
	Backward     key.Binding
	VolUp        key.Binding
	VolDown      key.Binding
	ToggleLyrics key.Binding
}

func newHelpMap() *helpKeyMap {
	controls := config.Current.Controls
	return &helpKeyMap{
		PlayPause: key.NewBinding(
			controls.PlayerPause.Binding(),
			controls.PlayerPause.Help("play/pause"),
		),
		PrevTrack: key.NewBinding(
			controls.PlayerPrevious.Binding(),
			controls.PlayerPrevious.Help("previous track"),
		),
		NextTrack: key.NewBinding(
			controls.PlayerNext.Binding(),
			controls.PlayerNext.Help("next track"),
		),
		LikeUnlike: key.NewBinding(
			controls.PlayerLike.Binding(),
			controls.PlayerLike.Help("like/unlike"),
		),
		CacheTrack: key.NewBinding(
			controls.PlayerCache.Binding(),
			controls.PlayerCache.Help("cache track"),
		),
		Backward: key.NewBinding(
			controls.PlayerRewindBackward.Binding(),
			controls.PlayerRewindBackward.Help(fmt.Sprintf("-%d sec", int(config.Current.RewindDuration))),
		),
		Forward: key.NewBinding(
			controls.PlayerRewindForward.Binding(),
			controls.PlayerRewindForward.Help(fmt.Sprintf("+%d sec", int(config.Current.RewindDuration))),
		),
		VolUp: key.NewBinding(
			controls.PlayerVolUp.Binding(),
			controls.PlayerVolUp.Help("vol up"),
		),
		VolDown: key.NewBinding(
			controls.PlayerVolDown.Binding(),
			controls.PlayerVolDown.Help("vol down"),
		),
		ToggleLyrics: key.NewBinding(
			controls.PlayerToggleLyrics.Binding(),
			controls.PlayerToggleLyrics.Help("show/hide lyrics"),
		),
	}
}

func (k helpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.PlayPause, k.NextTrack, k.PrevTrack, k.LikeUnlike}
}

func (k helpKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.PlayPause, k.LikeUnlike, k.ToggleLyrics, k.CacheTrack},
		{k.NextTrack, k.PrevTrack, k.Forward, k.Backward},
		{k.VolUp, k.VolDown},
	}
}
