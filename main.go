package main

import (
	"github.com/dece2183/yamusic-tui/log"
	"github.com/dece2183/yamusic-tui/ui"
)

func main() {
	log.Start()
	ui.Run()
	log.Stop()
}
