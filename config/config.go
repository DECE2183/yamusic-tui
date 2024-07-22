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

	configDir := filepath.Join(userDir, ".config", configPath)
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

	newControls := reflect.ValueOf(&newConfig.Controls).Elem()
	defaultControls := reflect.ValueOf(defaultConfig.Controls)
	for i := 0; i < newControls.NumField(); i++ {
		newField := newControls.Field(i)
		defaultField := defaultControls.Field(i)
		if newField.Len() == 0 {
			newField.SetString(defaultField.String())
		}
	}

	if newConfig.Controls.Quit == "" {
		newConfig.Controls.Quit = defaultConfig.Controls.Quit
	}

	if newConfig.VolumeStep == 0 {
		newConfig.VolumeStep = defaultConfig.VolumeStep
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
