package main

import (
	"runtime"

	"github.com/dece2183/yamusic-tui/log"
	"github.com/dece2183/yamusic-tui/media/handler/macos"
	"github.com/dece2183/yamusic-tui/ui"
)

func init() {
	// Lock the main goroutine to the main OS thread so that
	// the Cocoa run loop (which requires the main thread) can be started here.
	runtime.LockOSThread()
}

func main() {
	log.Start()

	go func() {
		ui.Run()
		log.Stop()
		macos.StopMain()
	}()

	// Block the main OS thread running the Cocoa event loop.
	// This is required for MPRemoteCommandCenter and NSEvent monitoring to work.
	macos.RunMain()
}
