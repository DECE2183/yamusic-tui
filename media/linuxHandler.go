//go:build linux && !nomedia

package media

import (
	"github.com/bircoder432/yamusic-tui-enhanced/media/handler"
	"github.com/bircoder432/yamusic-tui-enhanced/media/handler/mpris"
)

func NewHandler(name, description string) handler.MediaHandler {
	return mpris.NewHandler(name, description)
}
