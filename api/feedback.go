package api

import "fmt"

const (
	_FEEDBACK_FROM_FIELD = "web-home-rup_main-radio-default"
)

type RadioEventType string

const (
	EV_RADIO_STARTED  = RadioEventType(ROTOR_RADIO_STARTED)
	EV_RADIO_FINISHED = RadioEventType(ROTOR_RADIO_FINISHED)
)

type TrackEventType string

const (
	EV_TRACK_STARTED  = TrackEventType(ROTOR_TRACK_STARTED)
	EV_TRACK_FINISHED = TrackEventType(ROTOR_TRACK_FINISHED)
	EV_TRACK_SKIPED   = TrackEventType(ROTOR_SKIP)
	EV_TRACK_LIKED    = TrackEventType(ROTOR_LIKE)
	EV_TRACK_UNLIKED  = TrackEventType(ROTOR_UNLIKE)
)

type RotorFeedbackEvent struct {
	Timestamp          string  `json:"timestamp"`
	Type               string  `json:"type"`
	From               string  `json:"from,omitempty"`
	TotalPlayedSeconds float64 `json:"totalPlayedSeconds,omitempty"`
	TrackLengthSeconds float64 `json:"trackLengthSeconds,omitempty"`
	TrackId            string  `json:"trackId,omitempty"`
}

type RotorFeedback struct {
	From    string              `json:"from"`
	BatchId string              `json:"batchId,omitempty"`
	Event   *RotorFeedbackEvent `json:"event"`
}

func NewRadioFeedbackEvent(evType RadioEventType) *RotorFeedbackEvent {
	return &RotorFeedbackEvent{
		Timestamp: nowTimestamp(),
		From:      _FEEDBACK_FROM_FIELD,
		Type:      string(evType),
	}
}

func NewTrackFeedbackEvent(evType TrackEventType, track *Track, playedSeconds float64) *RotorFeedbackEvent {
	return &RotorFeedbackEvent{
		Timestamp:          nowTimestamp(),
		TotalPlayedSeconds: playedSeconds,
		TrackLengthSeconds: float64(track.DurationMs) * 1000.0,
		TrackId:            fmt.Sprintf("%s:%d", track.Id, track.Albums[0].Id),
		Type:               string(evType),
	}
}

func NewFeedback(batchId string, ev *RotorFeedbackEvent) *RotorFeedback {
	return &RotorFeedback{
		From:    _FEEDBACK_FROM_FIELD,
		BatchId: batchId,
		Event:   ev,
	}
}
