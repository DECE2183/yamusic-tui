package main

import (
	"fmt"
	"net/url"
	"os"
	"yamusic/api"
	"yamusic/config"
	"yamusic/ui"

	"github.com/charmbracelet/lipgloss"
)

func main() {
	cl, err := api.NewClient(config.GetToken())
	if err != nil {
		errMsg := lipgloss.NewStyle().Foreground(lipgloss.Color("#F33")).Render("Error:")
		if _, ok := err.(*url.Error); ok {
			fmt.Print("\n", errMsg, "unable to connect to the Yandex server\n\n")
		} else {
			fmt.Println(errMsg, err)
		}
		os.Exit(8)
	}

	ui.Run(cl)
}
