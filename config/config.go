package config

import (
	"os"
	"path/filepath"
)

var userToken string

func init() {
	userToken = LoadToken()
}

func GetToken() string {
	return userToken
}

func SaveToken(token string) error {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(userDir, ".config/yamusic-tui")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(filepath.Join(configDir, "token"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(token))
	if err != nil {
		return err
	}

	userToken = token
	return nil
}

func LoadToken() string {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	configDir := filepath.Join(userDir, ".config/yamusic-tui")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return ""
	}

	token, err := os.ReadFile(filepath.Join(configDir, "token"))
	if err != nil {
		return ""
	}

	return string(token)
}
