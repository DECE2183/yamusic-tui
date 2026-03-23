package config

import "gopkg.in/yaml.v3"

type CacheType uint

const (
	CACHE_NONE CacheType = iota
	CACHE_LIKED_ONLY
	CACHE_ALL
)

var cacheValueToEnum = map[string]CacheType{
	"disable": CACHE_NONE,
	"false":   CACHE_NONE,
	"none":    CACHE_NONE,
	"off":     CACHE_NONE,
	"likes":   CACHE_LIKED_ONLY,
	"liked":   CACHE_LIKED_ONLY,
	"all":     CACHE_ALL,
}

var cacheEnumToValue = map[CacheType]string{
	CACHE_NONE:       "none",
	CACHE_LIKED_ONLY: "likes",
	CACHE_ALL:        "all",
}

func (t *CacheType) UnmarshalYAML(value *yaml.Node) error {
	*t = cacheValueToEnum[value.Value]
	return nil
}

func (t CacheType) MarshalYAML() (interface{}, error) {
	if t > CACHE_ALL {
		t = CACHE_NONE
	}
	return cacheEnumToValue[t], nil
}

type Icons struct {
	Play      string `yaml:"play"`
	Stop      string `yaml:"stop"`
	Liked     string `yaml:"liked"`
	NotLiked  string `yaml:"not-liked"`
	Cached    string `yaml:"cached"`
	LyricsDot string `yaml:"lyrics-dot"`
}

type Colors struct {
	Accent            string `yaml:"accent"`
	Error             string `yaml:"error"`
	Border            string `yaml:"border"`
	Background        string `yaml:"background"`
	PlaylistSelection string `yaml:"playlist-selection"`
	ActiveText        string `yaml:"active-text"`
	NormalText        string `yaml:"normal-text"`
	InactiveText      string `yaml:"inactive-text"`
	TrackTitleText    string `yaml:"track-title-text"`
	TrackVersionText  string `yaml:"track-version-text"`
	TrackArtistText   string `yaml:"track-artist-text"`
	LyricsPrevious    string `yaml:"lyrics-previous"`
	LyricsCurrent     string `yaml:"lyrics-current"`
	LyricsNext        string `yaml:"lyrics-next"`
}

type Style struct {
	SidePanelWidth   int     `yaml:"side-panel-width"`
	SearchModalWidth int     `yaml:"search-modal-width"`
	Icons            *Icons  `yaml:"icons"`
	Colors           *Colors `yaml:"colors"`
}

type Controls struct {
	// Main control
	Quit        *Key `yaml:"quit"`
	Apply       *Key `yaml:"apply"`
	Cancel      *Key `yaml:"cancel"`
	CursorUp    *Key `yaml:"cursor-up"`
	CursorDown  *Key `yaml:"cursor-down"`
	Reload      *Key `yaml:"reload"`
	ShowAllKeys *Key `yaml:"show-all-kyes"`
	// Playlists control
	PlaylistsUp     *Key `yaml:"playlists-up"`
	PlaylistsDown   *Key `yaml:"playlists-down"`
	PlaylistsRename *Key `yaml:"playlists-rename"`
	PlaylistsHide   *Key `yaml:"playlists-hide"`
	// Track list control
	PageUp                   *Key `yaml:"page-up"`
	PageDown                 *Key `yaml:"page-down"`
	TracksLike               *Key `yaml:"tracks-like"`
	TracksAddToPlaylist      *Key `yaml:"tracks-add-to-playlist"`
	TracksRemoveFromPlaylist *Key `yaml:"tracks-remove-from-playlist"`
	TracksShare              *Key `yaml:"tracks-share"`
	TracksShuffle            *Key `yaml:"tracks-shuffle"`
	TracksSearch             *Key `yaml:"tracks-search"`
	TracksHide               *Key `yaml:"tracks-hide"`
	// Player control
	PlayerPause          *Key `yaml:"player-pause"`
	PlayerNext           *Key `yaml:"player-next"`
	PlayerPrevious       *Key `yaml:"player-previous"`
	PlayerRewindForward  *Key `yaml:"player-rewind-forward"`
	PlayerRewindBackward *Key `yaml:"player-rewind-backward"`
	PlayerLike           *Key `yaml:"player-like"`
	PlayerCache          *Key `yaml:"player-cache"`
	PlayerVolUp          *Key `yaml:"player-vol-up"`
	PlayerVolDown        *Key `yaml:"player-vol-down"`
	PlayerToggleLyrics   *Key `yaml:"player-toggle-lyrics"`
	PlayerHide           *Key `yaml:"player-hide"`
}

type Search struct {
	Artists   bool `yaml:"artists"`
	Albums    bool `yaml:"albums"`
	Playlists bool `yaml:"playlists"`
}

type Config struct {
	Token          string    `yaml:"token"`
	BufferSize     float64   `yaml:"buffer-size-ms"`
	RewindDuration float64   `yaml:"rewind-duration-s"`
	Volume         float64   `yaml:"volume"`
	VolumeStep     float64   `yaml:"volume-step"`
	SuppressErrors bool      `yaml:"suppress-errors"`
	ShowLyrics     bool      `yaml:"show-lyrics"`
	CacheTracks    CacheType `yaml:"cache-tracks"`
	CacheDir       string    `yaml:"cache-dir"`
	Search         *Search   `yaml:"search"`
	Controls       *Controls `yaml:"controls"`
	Style          *Style    `yaml:"style"`
}

var defaultConfig = Config{
	BufferSize:     80,
	RewindDuration: 5,
	Volume:         0.5,
	VolumeStep:     0.05,
	ShowLyrics:     false,
	CacheTracks:    CACHE_LIKED_ONLY,
	CacheDir:       "",
	SuppressErrors: false,
	Search: &Search{
		Artists:   true,
		Albums:    false,
		Playlists: false,
	},
	Controls: &Controls{
		Quit:                     NewKey("ctrl+q,ctrl+c"),
		Apply:                    NewKey("enter"),
		Cancel:                   NewKey("esc"),
		CursorUp:                 NewKey("up"),
		CursorDown:               NewKey("down"),
		Reload:                   NewKey("ctrl+\\"),
		ShowAllKeys:              NewKey("?"),
		PlaylistsUp:              NewKey("ctrl+up"),
		PlaylistsDown:            NewKey("ctrl+down"),
		PlaylistsRename:          NewKey("ctrl+r"),
		PlaylistsHide:            NewKey("ctrl+b"),
		PageUp:                   NewKey("pgup"),
		PageDown:                 NewKey("pgdown"),
		TracksLike:               NewKey("l"),
		TracksAddToPlaylist:      NewKey("a"),
		TracksRemoveFromPlaylist: NewKey("ctrl+a"),
		TracksSearch:             NewKey("ctrl+f"),
		TracksShuffle:            NewKey("ctrl+x"),
		TracksShare:              NewKey("ctrl+s"),
		TracksHide:               NewKey("ctrl+t"),
		PlayerPause:              NewKey("space"),
		PlayerNext:               NewKey("right"),
		PlayerPrevious:           NewKey("left"),
		PlayerRewindForward:      NewKey("ctrl+right"),
		PlayerRewindBackward:     NewKey("ctrl+left"),
		PlayerLike:               NewKey("L"),
		PlayerToggleLyrics:       NewKey("t"),
		PlayerCache:              NewKey("S"),
		PlayerVolUp:              NewKey("+,="),
		PlayerVolDown:            NewKey("-"),
		PlayerHide:               NewKey("ctrl+p"),
	},
	Style: &Style{
		SidePanelWidth:   32,
		SearchModalWidth: 56,
		Icons: &Icons{
			Play:      "▶",
			Stop:      "■",
			Liked:     "💛",
			NotLiked:  "🤍",
			Cached:    "💿",
			LyricsDot: "•",
		},
		Colors: &Colors{
			Accent:            "#FC0",
			Error:             "#F33",
			Border:            "#444",
			Background:        "#6b6b6b",
			PlaylistSelection: "#4a3c00",
			ActiveText:        "#EEE",
			NormalText:        "#CCC",
			InactiveText:      "#888",
			TrackTitleText:    "#dcdcdc",
			TrackVersionText:  "#999",
			TrackArtistText:   "#bbb",
			LyricsPrevious:    "#444",
			LyricsCurrent:     "#EEE",
			LyricsNext:        "#777",
		},
	},
}

const DirName = "yamusic-tui"
