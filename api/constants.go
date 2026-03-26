package api

// Yandex OAuth service information
const (
	yaOauthServerURL    = "https://oauth.yandex.ru/"
	yaOauthClientID     = "23cabbbdc6cd418abb4b39c32c41195d"
	yaOauthClientSecret = "53bc75238f0c4d08a118e51fe9203300"
)

// Yandex Music service information
const (
	YaMusicServerURL = "https://api.music.yandex.net/"
)

// Rotor feedback types
const (
	ROTOR_RADIO_STARTED  string = "radioStarted"
	ROTOR_RADIO_FINISHED string = "radioFinished"
	ROTOR_TRACK_STARTED  string = "trackStarted"
	ROTOR_TRACK_FINISHED string = "trackFinished"
	ROTOR_SKIP           string = "skip"
	ROTOR_LIKE           string = "like"
	ROTOR_UNLIKE         string = "unlike"
)

var (
	MyWaveId = StationId{
		Type: "user",
		Tag:  "onyourwave",
	}
)
