package handler

import (
	"time"
)

type MessageType int

const (
	MSG_NONE MessageType = iota

	MSG_NEXT
	MSG_PREVIOUS
	MSG_PLAY
	MSG_PAUSE
	MSG_PLAYPAUSE
	MSG_STOP
	MSG_SEEK
	MSG_SETPOS

	MSG_GET_PLAYBACKSTATUS
	MSG_GET_SHUFFLE
	MSG_GET_METADATA
	MSG_GET_VOLUME
	MSG_GET_POSITION

	MSG_SET_SHUFFLE
	MSG_SET_VOLUME
)

type Message struct {
	Type MessageType
	Arg  any
}

type TrackMetadata struct {
	TrackId      string
	Title        string
	Url          string
	CoverUrl     string
	Length       time.Duration
	Genre        []string
	Artists      []string
	AlbumName    string
	AlbumArtists []string
}

type PlaybackState int

const (
	STATE_STOPED PlaybackState = iota
	STATE_PAUSED
	STATE_PLAYING
)

type MediaHandler interface {
	Enable() error
	Disable() error
	Message() <-chan Message
	SendAnswer(ans any)

	OnEnded()
	OnVolume()
	OnPlayback()
	OnPlayPause()
	OnSeek(position time.Duration)
}
