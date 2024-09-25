package tracker

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/dece2183/yamusic-tui/config"
)

type helpKeyMap struct {
	PlayPause  key.Binding
	PrevTrack  key.Binding
	NextTrack  key.Binding
	LikeUnlike key.Binding
	Forward    key.Binding
	Backward   key.Binding
	VolUp      key.Binding
	VolDown    key.Binding
}

var helpMap = helpKeyMap{
	PlayPause: key.NewBinding(
		config.Current.Controls.PlayerPause.Binding(),
		config.Current.Controls.PlayerPause.Help("play/pause"),
	),
	PrevTrack: key.NewBinding(
		config.Current.Controls.PlayerPrevious.Binding(),
		config.Current.Controls.PlayerPrevious.Help("previous track"),
	),
	NextTrack: key.NewBinding(
		config.Current.Controls.PlayerNext.Binding(),
		config.Current.Controls.PlayerNext.Help("next track"),
	),
	LikeUnlike: key.NewBinding(
		config.Current.Controls.PlayerLike.Binding(),
		config.Current.Controls.PlayerLike.Help("like/unlike"),
	),
	Backward: key.NewBinding(
		config.Current.Controls.PlayerRewindBackward.Binding(),
		config.Current.Controls.PlayerRewindBackward.Help(fmt.Sprintf("-%d sec", int(config.Current.RewindDuration))),
	),
	Forward: key.NewBinding(
		config.Current.Controls.PlayerRewindForward.Binding(),
		config.Current.Controls.PlayerRewindForward.Help(fmt.Sprintf("+%d sec", int(config.Current.RewindDuration))),
	),
	VolUp: key.NewBinding(
		config.Current.Controls.PlayerVolUp.Binding(),
		config.Current.Controls.PlayerVolUp.Help("vol up"),
	),
	VolDown: key.NewBinding(
		config.Current.Controls.PlayerVolDown.Binding(),
		config.Current.Controls.PlayerVolDown.Help("vol down"),
	),
}

func (k helpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.PlayPause, k.NextTrack, k.PrevTrack, k.LikeUnlike}
}

func (k helpKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.PlayPause, k.LikeUnlike},
		{k.NextTrack, k.PrevTrack},
		{k.Forward, k.Backward},
		{k.VolUp, k.VolDown},
	}
}
