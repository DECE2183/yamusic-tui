package main

import (
	"yamusic/api"
	"yamusic/config"
	"yamusic/ui"
)

func main() {
	cl, err := api.NewClient(config.GetToken())
	if err != nil {
		panic(err)
	}

	ui.Run(cl)
}
