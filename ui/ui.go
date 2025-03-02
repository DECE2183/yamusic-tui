package ui

import (
	"github.com/dece2183/yamusic-tui/config"
	"github.com/dece2183/yamusic-tui/log"
	"github.com/dece2183/yamusic-tui/ui/model"
	loginpage "github.com/dece2183/yamusic-tui/ui/model/loginPage"
	mainpage "github.com/dece2183/yamusic-tui/ui/model/mainPage"
)

func Run() {
	var err error

	if config.Current.Token == "" {
		err = loginpage.New().Run()
		if err != nil {
			log.Print(log.LVL_PANIC, err.Error())
			model.PrettyExit(err, 4)
		}
	}

	err = mainpage.New().Run()
	if err != nil {
		log.Print(log.LVL_PANIC, err.Error())
		model.PrettyExit(err, 6)
	}
}
