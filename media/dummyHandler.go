//go:build nomedia

package media

import (
	"github.com/dece2183/yamusic-tui/media/handler"
	"github.com/dece2183/yamusic-tui/media/handler/dummy"
)

func NewHandler(name, description string) handler.MediaHandler {
	return dummy.NewHandler(name, description)
}
