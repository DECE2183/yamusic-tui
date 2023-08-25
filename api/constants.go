package api

// Yandex OAuth service information
const (
	yaOauthServerURL    = "https://oauth.yandex.ru/"
	yaOauthClientID     = "23cabbbdc6cd418abb4b39c32c41195d"
	yaOauthClientSecret = "53bc75238f0c4d08a118e51fe9203300"
)

// Yandex Music service information
const (
	YaMusicServerURL = "https://api.music.yandex.net:443/"
)

// Rotor feedback types
const (
	ROTOR_RADIO_STARTED  = "radioStarted"
	ROTOR_RADIO_FINISHED = "radioFinished"
	ROTOR_TRACK_STARTED  = "trackStarted"
	ROTOR_TRACK_FINISHED = "trackFinished"
	ROTOR_SKIP           = "skip"
)

var (
	MyWaveId = StationId{
		Type: "user",
		Tag:  "onyourwave",
	}
)
