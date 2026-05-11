package cache

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/config"
)

type PlaylistEntry struct {
	Playlist api.Playlist `json:"playlist"`
	Tracks   []api.Track  `json:"tracks"`
}

type PlaylistsData struct {
	Entries []PlaylistEntry `json:"entries"`
}

func playlistsPath() (string, error) {
	userDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userDir, config.DirName, "playlists.json"), nil
}

func ReadPlaylists() (*PlaylistsData, error) {
	p, err := playlistsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var d PlaylistsData
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

func WritePlaylists(d *PlaylistsData) error {
	p, err := playlistsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	data, err := json.Marshal(d)
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}
