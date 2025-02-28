package ui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	loginpage "github.com/dece2183/yamusic-tui/ui/model/loginPage"
	mainpage "github.com/dece2183/yamusic-tui/ui/model/mainPage"
	"os"

	"github.com/dece2183/yamusic-tui/config"
	"github.com/dece2183/yamusic-tui/ui/model"
)

func Run() {
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}
	var err error

	if config.Current.Token == "" {
		err = loginpage.New().Run()
		if err != nil {
			model.PrettyExit(err, 4)
		}
	}

	err = mainpage.New().Run()
	if err != nil {
		model.PrettyExit(err, 6)
	}
}
