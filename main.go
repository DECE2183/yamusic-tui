package main

import (
	"github.com/bircoder432/yamusic-tui-enhanced/log"
	"github.com/bircoder432/yamusic-tui-enhanced/ui"
)

func main() {
	log.Start()
	ui.Run()
	log.Stop()
}
