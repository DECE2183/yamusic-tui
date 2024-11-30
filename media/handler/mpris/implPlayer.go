//go:build linux && !nomedia

package mpris

import (
	"fmt"
	"time"

	"github.com/dece2183/yamusic-tui/media/handler"
	"github.com/godbus/dbus/v5"
	"github.com/quarckster/go-mpris-server/pkg/types"
)

// MediaPlayer2.Player dbus interface implementation

func (mh *MprisHandler) Next() error {
	mh.msgChan <- handler.Message{
		Type: handler.MSG_NEXT,
	}
	return nil
}

func (mh *MprisHandler) Previous() error {
	mh.msgChan <- handler.Message{
		Type: handler.MSG_PREVIOUS,
	}
	return nil
}

func (mh *MprisHandler) Pause() error {
	mh.msgChan <- handler.Message{
		Type: handler.MSG_PAUSE,
	}
	return nil
}

func (mh *MprisHandler) PlayPause() error {
	mh.msgChan <- handler.Message{
		Type: handler.MSG_PLAYPAUSE,
	}
	return nil
}

func (mh *MprisHandler) Stop() error {
	mh.msgChan <- handler.Message{
		Type: handler.MSG_STOP,
	}
	return nil
}

func (mh *MprisHandler) Play() error {
	mh.msgChan <- handler.Message{
		Type: handler.MSG_PLAY,
	}
	return nil
}

func (mh *MprisHandler) Seek(offset types.Microseconds) error {
	mh.msgChan <- handler.Message{
		Type: handler.MSG_SEEK,
		Arg:  time.Duration(offset) * time.Microsecond,
	}
	return nil
}

func (mh *MprisHandler) SetPosition(trackId string, position types.Microseconds) error {
	mh.msgChan <- handler.Message{
		Type: handler.MSG_GET_METADATA,
	}

	resp, ok := (<-mh.ansChan).(handler.TrackMetadata)
	if !ok || resp.TrackId != trackId {
		return fmt.Errorf("trackId mismatch")
	}

	mh.msgChan <- handler.Message{
		Type: handler.MSG_SETPOS,
		Arg:  time.Duration(position) * time.Microsecond,
	}

	return nil
}

func (mh *MprisHandler) OpenUri(uri string) error {
	return nil
}

func (mh *MprisHandler) PlaybackStatus() (types.PlaybackStatus, error) {
	mh.msgChan <- handler.Message{
		Type: handler.MSG_GET_PLAYBACKSTATUS,
	}

	resp, ok := (<-mh.ansChan).(handler.PlaybackState)
	if !ok {
		return types.PlaybackStatusStopped, fmt.Errorf("wrong playback status type")
	}

	switch resp {
	case handler.STATE_STOPED:
		return types.PlaybackStatusStopped, nil
	case handler.STATE_PAUSED:
		return types.PlaybackStatusPaused, nil
	case handler.STATE_PLAYING:
		return types.PlaybackStatusPlaying, nil
	}

	return types.PlaybackStatusStopped, fmt.Errorf("unknown playback status")
}

func (mh *MprisHandler) Rate() (float64, error) {
	return 1, nil
}

func (mh *MprisHandler) SetRate(float64) error {
	return nil
}

func (mh *MprisHandler) Metadata() (md types.Metadata, err error) {
	mh.msgChan <- handler.Message{
		Type: handler.MSG_GET_METADATA,
	}

	resp, ok := (<-mh.ansChan).(handler.TrackMetadata)
	if !ok {
		err = fmt.Errorf("wrong metadata type")
		return
	}

	if len(resp.TrackId) == 0 {
		resp.TrackId = "NoTrack"
	}

	md.TrackId = dbus.ObjectPath("/org/mpris/MediaPlayer2/" + resp.TrackId)
	md.Length = types.Microseconds(resp.Length.Microseconds())
	md.ArtUrl = resp.CoverUrl
	md.Album = resp.AlbumName
	md.AlbumArtist = resp.AlbumArtists
	md.Artist = resp.Artists
	md.Genre = resp.Genre
	md.Title = resp.Title
	md.Url = resp.Url

	return
}

func (mh *MprisHandler) Volume() (float64, error) {
	mh.msgChan <- handler.Message{
		Type: handler.MSG_GET_VOLUME,
	}

	resp, ok := (<-mh.ansChan).(float64)
	if !ok {
		return 0, fmt.Errorf("wrong volume type")
	}

	return resp, nil
}

func (mh *MprisHandler) SetVolume(vol float64) error {
	mh.msgChan <- handler.Message{
		Type: handler.MSG_SET_VOLUME,
		Arg:  vol,
	}
	return nil
}

func (mh *MprisHandler) Position() (int64, error) {
	mh.msgChan <- handler.Message{
		Type: handler.MSG_GET_POSITION,
	}

	resp, ok := (<-mh.ansChan).(time.Duration)
	if !ok {
		return 0, fmt.Errorf("wrong position type")
	}

	return resp.Microseconds(), nil
}

func (mh *MprisHandler) MinimumRate() (float64, error) {
	return 1, nil
}

func (mh *MprisHandler) MaximumRate() (float64, error) {
	return 1, nil
}

func (mh *MprisHandler) CanGoNext() (bool, error) {
	return true, nil
}

func (mh *MprisHandler) CanGoPrevious() (bool, error) {
	return true, nil
}

func (mh *MprisHandler) CanPlay() (bool, error) {
	return true, nil
}

func (mh *MprisHandler) CanPause() (bool, error) {
	return true, nil
}

func (mh *MprisHandler) CanSeek() (bool, error) {
	return true, nil
}

func (mh *MprisHandler) CanControl() (bool, error) {
	return true, nil
}
