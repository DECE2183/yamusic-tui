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
	return &helpKeyMap{
		apply: key.NewBinding(
			config.Current.Controls.Apply.Binding(),
			config.Current.Controls.Apply.Help("apply"),
		),
		cancel: key.NewBinding(
			config.Current.Controls.Cancel.Binding(),
			config.Current.Controls.Cancel.Help("cancel"),
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
