package config

import (
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"gopkg.in/yaml.v3"
)

type Key struct {
	displayName string
	keyNames    []string
}

func NewKey(key string) *Key {
	k := &Key{
		displayName: prepareToDisplay(key),
		keyNames:    prepareToProccess(key),
	}
	return k
}

func (k *Key) IsEmpty() bool {
	return k == nil || len(k.keyNames) == 0
}

func (k *Key) Binding() key.BindingOpt {
	return key.WithKeys(k.keyNames...)
}

func (k *Key) Help(help string) key.BindingOpt {
	return key.WithHelp(k.displayName, help)
}

func (k *Key) Contains(keyName string) bool {
	return slices.Contains(k.keyNames, keyName)
}

func (k *Key) MarshalYAML() (interface{}, error) {
	return strings.Join(k.keyNames, ","), nil
}

func (k *Key) UnmarshalYAML(val *yaml.Node) error {
	k.displayName = prepareToDisplay(val.Value)
	k.keyNames = prepareToProccess(val.Value)
	return nil
}

func prepareToProccess(key string) []string {
	names := make([]string, 0)
	for _, part := range strings.Split(key, ",") {
		if part == "backspace" {
			// contains the substring "space", which must not be substituted.
			names = append(names, "backspace")
			continue
		}
		s := strings.ReplaceAll(part, "space", " ")
		s = strings.ReplaceAll(s, "↑", "up")
		s = strings.ReplaceAll(s, "↓", "down")
		s = strings.ReplaceAll(s, "←", "left")
		s = strings.ReplaceAll(s, "→", "right")
		names = append(names, s)
	}
	return names
}

func prepareToDisplay(key string) string {
	var s = strings.ReplaceAll(key, " ", "space")
	s = strings.ReplaceAll(s, "up", "↑")
	s = strings.ReplaceAll(s, "down", "↓")
	s = strings.ReplaceAll(s, "left", "←")
	s = strings.ReplaceAll(s, "right", "→")
	return s
}
