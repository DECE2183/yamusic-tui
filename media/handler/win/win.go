//go:build windows

package win

import (
	"time"

	"github.com/dece2183/yamusic-tui/media/handler"
)

type WinHandler struct {
	msgChan chan handler.Message
	ansChan chan any
}

func NewHandler(name, description string) *WinHandler {
	return &WinHandler{
		msgChan: make(chan handler.Message),
		ansChan: make(chan any),
	}
}

func (mh *WinHandler) Enable() error {
	return nil
}

func (mh *WinHandler) Disable() error {
	return nil
}

func (mh *WinHandler) Message() <-chan handler.Message {
	return mh.msgChan
}

func (mh *WinHandler) SendAnswer(ans any) {
	mh.ansChan <- ans
}

func (mh *WinHandler) OnEnded() {
}

func (mh *WinHandler) OnVolume() {
}

func (mh *WinHandler) OnPlayback() {
}

func (mh *WinHandler) OnPlayPause() {
}

func (mh *WinHandler) OnSeek(position time.Duration) {
}
