package config

import (
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
)

type Key string

func (k Key) prepareToProccess() string {
	var s = strings.ReplaceAll(string(k), "space", " ")
	s = strings.ReplaceAll(s, "↑", "up")
	s = strings.ReplaceAll(s, "↓", "down")
	s = strings.ReplaceAll(s, "←", "left")
	s = strings.ReplaceAll(s, "→", "right")
	return s
}

func (k Key) prepareToDisplay() string {
	var s = strings.ReplaceAll(string(k), " ", "space")
	s = strings.ReplaceAll(s, "up", "↑")
	s = strings.ReplaceAll(s, "down", "↓")
	s = strings.ReplaceAll(s, "left", "←")
	s = strings.ReplaceAll(s, "right", "→")
	return s
}

func (k Key) Key() string {
	return k.prepareToProccess()
}

func (k Key) Binding() key.BindingOpt {
	s := k.prepareToProccess()
	keys := strings.Split(s, ",")
	return key.WithKeys(keys...)
}

func (k Key) Help(help string) key.BindingOpt {
	s := k.prepareToDisplay()
	return key.WithHelp(s, help)
}

func (k Key) Contains(keyName string) bool {
	s := k.prepareToProccess()
	keys := strings.Split(s, ",")
	return slices.Contains(keys, keyName)
}

type Controls struct {
	// Main control
	Quit Key `yaml:"quit"`
	// Playlists control
	PlaylistsUp   Key `yaml:"playlists-up"`
	PlaylistsDown Key `yaml:"playlists-down"`
	// Track list control
	TrackListUp     Key `yaml:"track-list-up"`
	TrackListDown   Key `yaml:"track-list-down"`
	TrackListSelect Key `yaml:"track-list-select"`
	TrackListLike   Key `yaml:"track-list-like"`
	TrackListShare  Key `yaml:"track-list-share"`
	// Player control
	PlayerPause          Key `yaml:"player-pause"`
	PlayerNext           Key `yaml:"player-next"`
	PlayerPrevious       Key `yaml:"player-previous"`
	PlayerRewindForward  Key `yaml:"player-rewind-forward"`
	PlayerRewindBackward Key `yaml:"player-rewind-backward"`
	PlayerLike           Key `yaml:"player-like"`
}

type Config struct {
	Token          string   `yaml:"token"`
	BufferSize     float64  `yaml:"buffer-size-ms"`
	RewindDuration float64  `yaml:"rewind-duration-s"`
	Volume         float64  `yaml:"volume"`
	Controls       Controls `yaml:"controls"`
}

var defaultConfig = Config{
	BufferSize:     80,
	RewindDuration: 5,
	Volume:         0.5,
	Controls: Controls{
		Quit:                 "ctrl+q,ctrl+c",
		PlaylistsUp:          "ctrl+up",
		PlaylistsDown:        "ctrl+down",
		TrackListUp:          "up",
		TrackListDown:        "down",
		TrackListSelect:      "enter",
		TrackListLike:        "l",
		TrackListShare:       "ctrl+s",
		PlayerPause:          "space",
		PlayerNext:           "right",
		PlayerPrevious:       "left",
		PlayerRewindForward:  "ctrl+right",
		PlayerRewindBackward: "ctrl+left",
		PlayerLike:           "L",
	},
}

const configPath = "yamusic-tui"
