package cache

import (
	"os"
	"path/filepath"

	"github.com/dece2183/yamusic-tui/config"
)

func getCacheDir() (string, error) {
	userDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	var cacheDir string
	if config.Current.CacheDir == "default" {
		cacheDir = filepath.Join(userDir, config.ConfigPath)
	} else {
		cacheDir, err = filepath.Abs(config.Current.CacheDir)
	}
	if err != nil {
		return "", err
	}
	err = os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return "", err
	}

	return cacheDir, nil
}

func Read(trackId string) (*os.File, int64, error) {
	dir, err := getCacheDir()
	if err != nil {
		return nil, 0, err
	}

	file, err := os.Open(filepath.Join(dir, trackId+".mp3"))
	if err != nil {
		return nil, 0, err
	}

	stat, _ := file.Stat()
	return file, stat.Size(), nil
}

func Write(trackId string) (*os.File, error) {
	dir, err := getCacheDir()
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filepath.Join(dir, trackId+".mp3"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func Remove(trackId string) error {
	dir, err := getCacheDir()
	if err != nil {
		return err
	}

	return os.Remove(filepath.Join(dir, trackId+".mp3"))
}
