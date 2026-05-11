package cache

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/config"
)

type LikedTracksData struct {
	Ids    []string    `json:"ids"`
	Tracks []api.Track `json:"tracks"`
}

func likedTracksPath() (string, error) {
	userDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userDir, config.DirName, "liked_tracks.json"), nil
}

func ReadLikedTracks() (*LikedTracksData, error) {
	p, err := likedTracksPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var c LikedTracksData
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func WriteLikedTracks(c *LikedTracksData) error {
	p, err := likedTracksPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}
