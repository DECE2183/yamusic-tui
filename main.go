package main

import (
	"github.com/dece2183/yamusic-tui/config"
	"github.com/dece2183/yamusic-tui/log"
	"github.com/dece2183/yamusic-tui/media"
	"github.com/dece2183/yamusic-tui/ui/model"
	loginpage "github.com/dece2183/yamusic-tui/ui/model/loginPage"
	mainpage "github.com/dece2183/yamusic-tui/ui/model/mainPage"
	"github.com/dece2183/yamusic-tui/ui/style"
)

func main() {
	log.Start()
	defer log.Stop()

	err := config.InitialLoad()
	if err != nil {
		log.Print(log.LVL_WARNIGN, "config load error: %s", err.Error())
	}

	style.Apply(config.Current.Style)

	if config.Current.Token == "" {
		err = loginpage.New().Run()
		if err != nil {
			log.Print(log.LVL_PANIC, err.Error())
			model.PrettyExit(err, 4)
		}
	}

	mediaHandler := media.NewHandler(config.DirName, "Yandex music terminal client")
	page := mainpage.New(mediaHandler)
	err = mediaHandler.Start(page.Run)
	if err != nil {
		log.Print(log.LVL_PANIC, err.Error())
		model.PrettyExit(err, 6)
	}
}
