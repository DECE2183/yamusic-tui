package config

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
)

type Key string

func (k Key) Key() string {
	return strings.ReplaceAll(string(k), "space", " ")
}

func (k Key) Binding() key.BindingOpt {
	return key.WithKeys(k.Key())
}

func (k Key) Help(help string) key.BindingOpt {
	var s = strings.ReplaceAll(string(k), " ", "space")
	s = strings.ReplaceAll(s, "up", "↑")
	s = strings.ReplaceAll(s, "down", "↓")
	s = strings.ReplaceAll(s, "left", "←")
	s = strings.ReplaceAll(s, "right", "→")
	return key.WithHelp(s, help)
}

type Controls struct {
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
	Controls       Controls `yaml:"controls"`
}

var defaultConfig = Config{
	BufferSize:     80,
	RewindDuration: 5,
	Controls: Controls{
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
