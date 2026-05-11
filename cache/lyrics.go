package cache

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/dece2183/yamusic-tui/api"
)

func lyricsDir() (string, error) {
	base, err := getCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, ".lyrics")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func lyricsPath(trackId string) (string, error) {
	dir, err := lyricsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, trackId+".json"), nil
}

func ReadLyrics(trackId string) ([]api.LyricPair, error) {
	p, err := lyricsPath(trackId)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var lyrics []api.LyricPair
	if err := json.Unmarshal(data, &lyrics); err != nil {
		return nil, err
	}
	return lyrics, nil
}

func WriteLyrics(trackId string, lyrics []api.LyricPair) error {
	p, err := lyricsPath(trackId)
	if err != nil {
		return err
	}
	data, err := json.Marshal(lyrics)
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}
