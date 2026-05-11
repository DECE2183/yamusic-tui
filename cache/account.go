package cache

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/dece2183/yamusic-tui/config"
)

type AccountData struct {
	Uid uint64 `json:"uid"`
}

func accountPath() (string, error) {
	userDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userDir, config.DirName, "account.json"), nil
}

func ReadAccount() (*AccountData, error) {
	p, err := accountPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var a AccountData
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

func WriteAccount(a *AccountData) error {
	p, err := accountPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	data, err := json.Marshal(a)
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}
