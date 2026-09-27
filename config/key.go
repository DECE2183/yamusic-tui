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
	return slices.Contains(k.keyNames, normalizeLayout(keyName))
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

// normalizeLayout maps Cyrillic letters to the Latin letters on the same
// physical keys (ЙЦУКЕН layout), preserving case, so bindings are independent
// of the active keyboard layout. Non-Latin runes pass through unchanged; letter
// hotkeys therefore fire regardless of whether RU or EN layout is active.
func normalizeLayout(text string) string {
	var b strings.Builder
	for _, r := range text {
		lat, ok := cyrillicToLatin[r]
		if !ok {
			b.WriteRune(r)
			continue
		}
		b.WriteRune(lat)
	}
	return b.String()
}

// cyrillicToLatin maps Cyrillic letters (ЙЦУКЕН layout) to the Latin letter on
// the same physical key.
var cyrillicToLatin = map[rune]rune{
	'й': 'q', 'ц': 'w', 'у': 'e', 'к': 'r', 'е': 't', 'н': 'y',
	'г': 'u', 'ш': 'i', 'щ': 'o', 'з': 'p', 'х': '[', 'ъ': ']',
	'ф': 'a', 'ы': 's', 'в': 'd', 'а': 'f', 'п': 'g', 'р': 'h',
	'о': 'j', 'л': 'k', 'д': 'l', 'ж': ';', 'э': '\'',
	'я': 'z', 'ч': 'x', 'с': 'c', 'м': 'v', 'и': 'b', 'т': 'n',
	'ь': 'm', 'б': ',', 'ю': '.',
	'Й': 'q', 'Ц': 'w', 'У': 'e', 'К': 'r', 'Е': 't', 'Н': 'y',
	'Г': 'u', 'Ш': 'i', 'Щ': 'o', 'З': 'p', 'Х': '[', 'Ъ': ']',
	'Ф': 'a', 'Ы': 's', 'В': 'd', 'А': 'f', 'П': 'g', 'Р': 'h',
	'О': 'j', 'Л': 'k', 'Д': 'l', 'Ж': ';', 'Э': '\'',
	'Я': 'z', 'Ч': 'x', 'С': 'c', 'М': 'v', 'И': 'b', 'Т': 'n',
	'Ь': 'm', 'Б': ',', 'Ю': '.',
}
