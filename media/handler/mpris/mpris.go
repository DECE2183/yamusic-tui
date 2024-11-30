//go:build linux && !nomedia

package mpris

import (
	"time"

	"github.com/dece2183/yamusic-tui/media/handler"
	"github.com/quarckster/go-mpris-server/pkg/events"
	"github.com/quarckster/go-mpris-server/pkg/server"
	"github.com/quarckster/go-mpris-server/pkg/types"
)

type MprisHandler struct {
	server      *server.Server
	evHandler   *events.EventHandler
	name        string
	description string
	msgChan     chan handler.Message
	ansChan     chan any
}

func NewHandler(name, description string) *MprisHandler {
	mh := &MprisHandler{
		name:        name,
		description: description,
		msgChan:     make(chan handler.Message),
		ansChan:     make(chan any),
	}

	mh.server = server.NewServer(mh.name, mh, mh)
	mh.evHandler = events.NewEventHandler(mh.server)

	return mh
}

func (mh *MprisHandler) Enable() error {
	go mh.server.Listen()
	return nil
}

func (mh *MprisHandler) Disable() error {
	err := mh.server.Stop()
	close(mh.msgChan)
	close(mh.ansChan)
	return err
}

func (mh *MprisHandler) Message() <-chan handler.Message {
	return mh.msgChan
}

func (mh *MprisHandler) SendAnswer(ans any) {
	mh.ansChan <- ans
}

func (mh *MprisHandler) OnEnded() {
	mh.evHandler.Player.OnEnded()
}

func (mh *MprisHandler) OnVolume() {
	mh.evHandler.Player.OnVolume()
}

func (mh *MprisHandler) OnPlayback() {
	mh.evHandler.Player.OnPlayback()
}

func (mh *MprisHandler) OnPlayPause() {
	mh.evHandler.Player.OnPlayPause()
}

func (mh *MprisHandler) OnSeek(position time.Duration) {
	mh.evHandler.Player.OnSeek(types.Microseconds(position.Microseconds()))
}
