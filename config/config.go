package config

import (
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v3"
)

var Current Config

func init() {
	var err error
	Current, err = load()
	if err != nil {
		configDir, err := getDir()
		if err != nil {
			panic(err)
		}

		if oldToken, err := os.ReadFile(filepath.Join(configDir, "token")); err == nil {
			Current.Token = string(oldToken)
		}

		save(Current)
	}
}

func getDir() (string, error) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(userDir, ".config", ConfigPath)
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return "", err
	}

	return configDir, nil
}

func load() (Config, error) {
	configDir, err := getDir()
	if err != nil {
		return defaultConfig, err
	}

	configContent, err := os.ReadFile(filepath.Join(configDir, "config.yaml"))
	if err != nil {
		return defaultConfig, err
	}

	var newConfig Config
	err = yaml.Unmarshal(configContent, &newConfig)
	if err != nil {
		return defaultConfig, err
	}

	if newConfig.VolumeStep == 0 {
		newConfig.VolumeStep = defaultConfig.VolumeStep
	}

	if newConfig.Search == nil {
		search := *defaultConfig.Search
		newConfig.Search = &search
	}

	if newConfig.Controls == nil {
		controls := *defaultConfig.Controls
		newConfig.Controls = &controls
	} else {
		newControls := reflect.ValueOf(newConfig.Controls).Elem()
		defaultControls := reflect.ValueOf(defaultConfig.Controls).Elem()
		for i := 0; i < newControls.NumField(); i++ {
			newField := newControls.Field(i).Interface().(*Key)
			defaultField := defaultControls.Field(i)
			if newField.IsEmpty() {
				newControls.Field(i).Set(defaultField)
			}
		}
		if newConfig.Controls.Quit.IsEmpty() {
			newConfig.Controls.Quit = defaultConfig.Controls.Quit
		}
	}

	return newConfig, nil
}

func save(conf Config) error {
	configDir, err := getDir()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filepath.Join(configDir, "config.yaml"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := yaml.NewEncoder(file)
	enc.SetIndent(4)
	err = enc.Encode(conf)
	if err != nil {
		return err
	}

	return nil
}

func Save() error {
	return save(Current)
}

func Reset() error {
	var err error
	Current, err = load()
	return err
}
