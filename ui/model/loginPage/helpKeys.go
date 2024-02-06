package loginpage

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/dece2183/yamusic-tui/config"
)

type helpKeyMap struct {
	apply key.Binding
	quit  key.Binding
}

func (k helpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.apply, k.quit}
}

func (k helpKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}

var helpMap = helpKeyMap{
	apply: key.NewBinding(
		config.Current.Controls.Apply.Binding(),
		config.Current.Controls.Apply.Help("login"),
	),
	quit: key.NewBinding(
		config.Current.Controls.Quit.Binding(),
		config.Current.Controls.Quit.Help("quit"),
	),
}
