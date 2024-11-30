//go:build windows && !nomedia

package win

import (
	_ "embed"
	"time"
	"unsafe"

	winrt "github.com/dece2183/media-winrt-go"
	"github.com/dece2183/media-winrt-go/windows/foundation"
	"github.com/dece2183/media-winrt-go/windows/media"
	"github.com/dece2183/media-winrt-go/windows/media/playback"
	"github.com/dece2183/media-winrt-go/windows/storage"
	"github.com/dece2183/yamusic-tui/media/handler"
	"github.com/go-ole/go-ole"
)

type PlayState int

const (
	PLAY_CLOSED  PlayState = -1
	PLAY_STOPED  PlayState = 0
	PLAY_PAUSED  PlayState = 1
	PLAY_PLAYING PlayState = 2
)

func (wh *WinHandler) initSmtc() error {
	err := ole.RoInitialize(1)
	if err != nil {
		return err
	}

	wh.mediaPlayer, err = playback.NewMediaPlayer()
	if err != nil {
		return err
	}

	cmdManager, err := wh.mediaPlayer.GetCommandManager()
	if err != nil {
		return err
	}
	err = cmdManager.SetIsEnabled(false)
	if err != nil {
		return err
	}

	wh.smtc, err = wh.mediaPlayer.GetSystemMediaTransportControls()
	if err != nil {
		return err
	}
	err = wh.smtc.SetIsPlayEnabled(true)
	if err != nil {
		return err
	}
	err = wh.smtc.SetIsPauseEnabled(true)
	if err != nil {
		return err
	}
	err = wh.smtc.SetIsPreviousEnabled(true)
	if err != nil {
		return err
	}
	err = wh.smtc.SetIsNextEnabled(true)
	if err != nil {
		return err
	}
	err = wh.smtc.SetIsRewindEnabled(true)
	if err != nil {
		return err
	}

	wh.buttonPressedEvent = makeEventHandler(wh.onButtonPressed, media.SignatureSystemMediaTransportControls, media.SignatureSystemMediaTransportControlsButtonPressedEventArgs)
	wh.buttonPressedToken, err = wh.smtc.AddButtonPressed(wh.buttonPressedEvent)
	if err != nil {
		return err
	}

	wh.positionChangedEvent = makeEventHandler(wh.onPositionChanged, media.SignatureSystemMediaTransportControls, media.SignaturePlaybackPositionChangeRequestedEventArgs)
	wh.positionChangedToken, err = wh.smtc.AddPlaybackPositionChangeRequested(wh.positionChangedEvent)
	if err != nil {
		return err
	}

	return nil
}

func (wh *WinHandler) smtcDispose() error {
	wh.smtc.RemoveButtonPressed(wh.positionChangedToken)
	wh.smtc.RemoveButtonPressed(wh.buttonPressedToken)
	wh.positionChangedEvent.Release()
	wh.buttonPressedEvent.Release()
	return wh.mediaPlayer.Close()
}

func (wh *WinHandler) setState(state PlayState) error {
	switch state {
	case PLAY_STOPED:
		return wh.smtc.SetPlaybackStatus(media.MediaPlaybackStatusStopped)
	case PLAY_PAUSED:
		return wh.smtc.SetPlaybackStatus(media.MediaPlaybackStatusPaused)
	case PLAY_PLAYING:
		return wh.smtc.SetPlaybackStatus(media.MediaPlaybackStatusPlaying)
	default:
		return wh.smtc.SetPlaybackStatus(media.MediaPlaybackStatusClosed)
	}
}

func (wh *WinHandler) setMetadata(path string) error {
	updater, err := wh.smtc.GetDisplayUpdater()
	if err != nil {
		return err
	}

	asyncOp, err := storage.StorageFileGetFileFromPathAsync(path)
	if err != nil {
		return err
	}

	err = awaitAsyncOperation(asyncOp, storage.SignatureStorageFile)
	if err != nil {
		return err
	}

	filePtr, err := asyncOp.GetResults()
	if err != nil {
		return err
	}

	file := (*storage.StorageFile)(filePtr)
	asyncOp, err = updater.CopyFromFileAsync(media.MediaPlaybackTypeMusic, file)
	if err != nil {
		return err
	}

	err = awaitAsyncOperation(asyncOp, winrt.SignatureBool)
	if err != nil {
		return err
	}

	successPtr, err := asyncOp.GetResults()
	if err != nil {
		return err
	}

	success := (uintptr(successPtr) == 1)
	if success {
		err = updater.Update()
		if err != nil {
			return err
		}
	}

	return nil
}

func (wh *WinHandler) updateTimeLineProperties(trackDuration, currentPos time.Duration) error {
	props, err := media.NewSystemMediaTransportControlsTimelineProperties()
	if err != nil {
		return err
	}

	startTime := foundation.TimeSpan{Duration: 0}
	err = props.SetStartTime(startTime)
	if err != nil {
		return err
	}
	err = props.SetMinSeekTime(startTime)
	if err != nil {
		return err
	}

	endTime := foundation.TimeSpan{Duration: trackDuration.Nanoseconds() / 100}
	err = props.SetEndTime(endTime)
	if err != nil {
		return err
	}
	err = props.SetMaxSeekTime(endTime)
	if err != nil {
		return err
	}

	position := foundation.TimeSpan{Duration: currentPos.Nanoseconds() / 100}
	err = props.SetPosition(position)
	if err != nil {
		return err
	}

	return wh.smtc.UpdateTimelineProperties(props)
}

func (wh *WinHandler) onButtonPressed(instance *foundation.TypedEventHandler, sender unsafe.Pointer, args unsafe.Pointer) {
	buttonArgs := (*media.SystemMediaTransportControlsButtonPressedEventArgs)(args)
	button, err := buttonArgs.GetButton()
	if err != nil {
		return
	}

	wh.msgMux.Lock()

	switch button {
	case media.SystemMediaTransportControlsButtonPlay:
		wh.msgChan <- handler.Message{
			Type: handler.MSG_PLAY,
		}
	case media.SystemMediaTransportControlsButtonPause:
		wh.msgChan <- handler.Message{
			Type: handler.MSG_PAUSE,
		}
	case media.SystemMediaTransportControlsButtonStop:
		wh.msgChan <- handler.Message{
			Type: handler.MSG_STOP,
		}
	case media.SystemMediaTransportControlsButtonNext:
		wh.msgChan <- handler.Message{
			Type: handler.MSG_NEXT,
		}
	case media.SystemMediaTransportControlsButtonPrevious:
		wh.msgChan <- handler.Message{
			Type: handler.MSG_PREVIOUS,
		}
	}

	wh.msgMux.Unlock()
}

func (wh *WinHandler) onPositionChanged(instance *foundation.TypedEventHandler, sender unsafe.Pointer, args unsafe.Pointer) {
	positionArgs := (*media.PlaybackPositionChangeRequestedEventArgs)(args)
	_ = positionArgs

	pos, err := positionArgs.GetRequestedPlaybackPosition()
	if err != nil {
		return
	}

	wh.msgMux.Lock()

	wh.msgChan <- handler.Message{
		Type: handler.MSG_SETPOS,
		Arg:  time.Duration(pos.Duration) * 100 * time.Nanosecond,
	}

	wh.msgMux.Unlock()
}
