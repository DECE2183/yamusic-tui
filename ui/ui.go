package ui

import (
	"yamusic/config"
	"yamusic/ui/model"
	loginpage "yamusic/ui/model/loginPage"
	mainpage "yamusic/ui/model/mainPage"
)

func Run() {
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
