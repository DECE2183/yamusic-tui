package config

type Controls struct {
	// Main control
	Quit       *Key `yaml:"quit"`
	Apply      *Key `yaml:"apply"`
	Cancel     *Key `yaml:"cancel"`
	CursorUp   *Key `yaml:"cursor-up"`
	CursorDown *Key `yaml:"cursor-down"`
	// Playlists control
	PlaylistsUp   *Key `yaml:"playlists-up"`
	PlaylistsDown *Key `yaml:"playlists-down"`
	// Track list control
	TracksLike    *Key `yaml:"tracks-like"`
	TracksShare   *Key `yaml:"tracks-share"`
	TracksShuffle *Key `yaml:"tracks-shuffle"`
	TracksSearch  *Key `yaml:"tracks-search"`
	// Player control
	PlayerPause          *Key `yaml:"player-pause"`
	PlayerNext           *Key `yaml:"player-next"`
	PlayerPrevious       *Key `yaml:"player-previous"`
	PlayerRewindForward  *Key `yaml:"player-rewind-forward"`
	PlayerRewindBackward *Key `yaml:"player-rewind-backward"`
	PlayerLike           *Key `yaml:"player-like"`
	PlayerVolUp          *Key `yaml:"player-vol-up"`
	PlayerVolDown        *Key `yaml:"player-vol-donw"`
}

type Config struct {
	Token          string   `yaml:"token"`
	BufferSize     float64  `yaml:"buffer-size-ms"`
	RewindDuration float64  `yaml:"rewind-duration-s"`
	Volume         float64  `yaml:"volume"`
	VolumeStep     float64  `yaml:"volume-step"`
	Controls       Controls `yaml:"controls"`
}

var defaultConfig = Config{
	BufferSize:     80,
	RewindDuration: 5,
	Volume:         0.5,
	VolumeStep:     0.05,
	Controls: Controls{
		Quit:                 NewKey("ctrl+q,ctrl+c"),
		Apply:                NewKey("enter"),
		Cancel:               NewKey("esc"),
		CursorUp:             NewKey("up"),
		CursorDown:           NewKey("down"),
		PlaylistsUp:          NewKey("ctrl+up"),
		PlaylistsDown:        NewKey("ctrl+down"),
		TracksLike:           NewKey("l"),
		TracksSearch:         NewKey("ctrl+f"),
		TracksShuffle:        NewKey("ctrl+x"),
		TracksShare:          NewKey("ctrl+s"),
		PlayerPause:          NewKey("space"),
		PlayerNext:           NewKey("right"),
		PlayerPrevious:       NewKey("left"),
		PlayerRewindForward:  NewKey("ctrl+right"),
		PlayerRewindBackward: NewKey("ctrl+left"),
		PlayerLike:           NewKey("L"),
		PlayerVolUp:          NewKey("+,="),
		PlayerVolDown:        NewKey("-"),
	},
}

const configPath = "yamusic-tui"
