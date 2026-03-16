package input

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/dece2183/yamusic-tui/config"
)

type helpKeyMap struct {
	apply  key.Binding
	cancel key.Binding

	Action string
}

func newHelpMap() *helpKeyMap {
	controls := config.Current.Controls
	return &helpKeyMap{
		apply: key.NewBinding(
			controls.Apply.Binding(),
			controls.Apply.Help("apply"),
		),
		cancel: key.NewBinding(
			controls.Cancel.Binding(),
			controls.Cancel.Help("cancel"),
		),
	}
}

func (k *helpKeyMap) ShortHelp() []key.Binding {
	k.apply.SetHelp(k.apply.Help().Key, k.Action)
	return []key.Binding{k.apply, k.cancel}
}

func (k *helpKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}
