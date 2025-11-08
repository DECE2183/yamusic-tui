//go:build windows && !nomedia

package media

import (
	"github.com/bircoder432/yamusic-tui-enhanced/media/handler"
	"github.com/bircoder432/yamusic-tui-enhanced/media/handler/win"
)

func NewHandler(name, description string) handler.MediaHandler {
	return win.NewHandler(name, description)
}
