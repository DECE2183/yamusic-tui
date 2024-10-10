//go:build windows

package media

import (
	"github.com/dece2183/yamusic-tui/media/handler"
	"github.com/dece2183/yamusic-tui/media/handler/win"
)

func NewHandler(name, description string) handler.MediaHandler {
	return win.NewHandler(name, description)
}
