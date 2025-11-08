package ui

import (
	"github.com/bircoder432/yamusic-tui-enhanced/config"
	"github.com/bircoder432/yamusic-tui-enhanced/log"
	"github.com/bircoder432/yamusic-tui-enhanced/ui/model"
	loginpage "github.com/bircoder432/yamusic-tui-enhanced/ui/model/loginPage"
	mainpage "github.com/bircoder432/yamusic-tui-enhanced/ui/model/mainPage"
	"github.com/bircoder432/yamusic-tui-enhanced/ui/style"
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
	style.InitStyles()
	err = mainpage.New().Run()
	if err != nil {
		log.Print(log.LVL_PANIC, err.Error())
		model.PrettyExit(err, 6)
	}
}
