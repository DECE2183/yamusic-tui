package cache

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/config"
)

type MyWaveData struct {
	StationId    api.StationId `json:"station_id"`
	SessionId    string        `json:"session_id"`
	SessionBatch string        `json:"session_batch"`
	Track        api.Track     `json:"track"`
}

func myWavePath() (string, error) {
	userDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userDir, config.DirName, "mywave.json"), nil
}

func ReadMyWave() (*MyWaveData, error) {
	p, err := myWavePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var d MyWaveData
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	if d.Track.Id == "" {
		return nil, os.ErrNotExist
	}
	return &d, nil
}

func WriteMyWave(d *MyWaveData) error {
	p, err := myWavePath()
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
