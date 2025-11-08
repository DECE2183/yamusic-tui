//go:build nomedia || darwin

package media

import (
	"github.com/bircoder432/yamusic-tui-enhanced/media/handler"
	"github.com/bircoder432/yamusic-tui-enhanced/media/handler/dummy"
)

func NewHandler(name, description string) handler.MediaHandler {
	return dummy.NewHandler(name, description)
}
