package cache

import (
	"os"
	"path/filepath"
)

func CoverDir() (string, error) {
	base, err := getCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, ".covers")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func CoverPath(trackId string) (string, error) {
	dir, err := CoverDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, trackId+".jpg"), nil
}
