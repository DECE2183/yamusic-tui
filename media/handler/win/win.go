//go:build windows && !nomedia

//go:generate go run github.com/tc-hib/go-winres make

package win

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dece2183/media-winrt-go/windows/foundation"
	"github.com/dece2183/media-winrt-go/windows/media"
	"github.com/dece2183/media-winrt-go/windows/media/playback"
	"github.com/dece2183/yamusic-tui/config"
	"github.com/dece2183/yamusic-tui/media/handler"
)

const _TIMELINE_POLL_PERIOD_MS = 400

type WinHandler struct {
	msgMux  sync.Mutex
	msgChan chan handler.Message
	ansChan chan any

	mediaPlayer *playback.MediaPlayer
	smtc        *media.SystemMediaTransportControls

	buttonPressedEvent   *foundation.TypedEventHandler
	buttonPressedToken   foundation.EventRegistrationToken
	positionChangedEvent *foundation.TypedEventHandler
	positionChangedToken foundation.EventRegistrationToken

	trackDuration time.Duration
	playState     PlayState
}

func NewHandler(name, description string) *WinHandler {
	return &WinHandler{
		msgChan:   make(chan handler.Message),
		ansChan:   make(chan any),
		playState: PLAY_CLOSED,
	}
}

func (wh *WinHandler) Enable() error {
	err := wh.initSmtc()
	if err != nil {
		return err
	}

	go wh.updateTimeline()
	return nil
}

func (wh *WinHandler) Disable() error {
	close(wh.msgChan)
	close(wh.ansChan)
	return wh.smtcDispose()
}

func (wh *WinHandler) Message() <-chan handler.Message {
	return wh.msgChan
}

func (wh *WinHandler) SendAnswer(ans any) {
	wh.ansChan <- ans
}

func (wh *WinHandler) OnEnded() {
}

func (wh *WinHandler) OnVolume() {
}

func (wh *WinHandler) OnPlayback() {
	if wh.playState == PLAY_CLOSED {
		wh.smtc.SetIsEnabled(true)
	}

	wh.playState = PLAY_PLAYING
	wh.setState(wh.playState)
	wh.setMetadata(filepath.Join(os.TempDir(), config.ConfigPath, "metadata.mp3"))

	wh.msgMux.Lock()
	wh.msgChan <- handler.Message{
		Type: handler.MSG_GET_METADATA,
	}
	metadata, ok := (<-wh.ansChan).(handler.TrackMetadata)
	wh.msgMux.Unlock()

	if ok {
		wh.trackDuration = metadata.Length
		wh.updateTimeLineProperties(wh.trackDuration, 0)
	}
}

func (wh *WinHandler) OnPlayPause() {
	if wh.playState != PLAY_PLAYING {
		wh.playState = PLAY_PLAYING
	} else {
		wh.playState = PLAY_PAUSED
	}

	wh.setState(wh.playState)
}

func (wh *WinHandler) OnSeek(position time.Duration) {
	wh.updateTimeLineProperties(wh.trackDuration, position)
}

func (wh *WinHandler) updateTimeline() {
	periodTimer := time.NewTicker(_TIMELINE_POLL_PERIOD_MS * time.Millisecond)

	for {
		<-periodTimer.C

		wh.msgMux.Lock()
		wh.msgChan <- handler.Message{
			Type: handler.MSG_GET_POSITION,
		}
		resp, ok := <-wh.ansChan
		wh.msgMux.Unlock()

		if !ok {
			periodTimer.Stop()
			return
		}

		pos, ok := resp.(time.Duration)
		if !ok {
			continue
		}

		wh.updateTimeLineProperties(wh.trackDuration, pos)
	}
}
