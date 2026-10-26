package loginpage

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/dece2183/yamusic-tui/config"
)

type helpKeyMap struct {
	apply key.Binding
	copy  key.Binding
	open  key.Binding
	retry key.Binding
	quit  key.Binding
}

func newHelpMap() *helpKeyMap {
	controls := config.Current.Controls
	return &helpKeyMap{
		apply: key.NewBinding(
			controls.Apply.Binding(),
			controls.Apply.Help("login"),
		),
		copy: key.NewBinding(
			key.WithKeys("ctrl+y"),
			key.WithHelp("ctrl+y", "copy code"),
		),
		open: key.NewBinding(
			key.WithKeys("ctrl+o"),
			key.WithHelp("ctrl+o", "open link"),
		),
		retry: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "new code"),
		),
		quit: key.NewBinding(
			controls.Quit.Binding(),
			controls.Quit.Help("quit"),
		),
	}
}

func (k helpKeyMap) ShortHelp() []key.Binding {
	if config.Current.DeviceLogin {
		return []key.Binding{k.copy, k.open, k.retry, k.quit}
	}
	return []key.Binding{k.apply, k.quit}
}

func (k helpKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}
