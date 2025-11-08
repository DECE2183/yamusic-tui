//go:build nomedia || darwin

package dummy

import (
	"time"

	"github.com/bircoder432/yamusic-tui-enhanced/media/handler"
)

type DummyHandler struct {
	msgChan chan handler.Message
}

func NewHandler(name, description string) *DummyHandler {
	return &DummyHandler{
		msgChan: make(chan handler.Message),
	}
}

func (*DummyHandler) Enable() error {
	return nil
}

func (dh *DummyHandler) Disable() error {
	close(dh.msgChan)
	return nil
}

func (dh *DummyHandler) Message() <-chan handler.Message {
	return dh.msgChan
}

func (*DummyHandler) SendAnswer(ans any) {
}

func (*DummyHandler) OnEnded() {
}

func (*DummyHandler) OnVolume() {
}

func (*DummyHandler) OnPlayback() {
}

func (*DummyHandler) OnPlayPause() {
}

func (*DummyHandler) OnSeek(position time.Duration) {
}
