package config

import "sync"

// radioMu guards RadioSettings, which is read from background goroutines (a
// tea.Cmd runs rotorSettingsFor off the Bubble Tea goroutine) while the Update
// goroutine may write it on APPLY. Protects against "concurrent map read and
// map write" panics.
var radioMu sync.RWMutex

// RadioSettings is read from background goroutines (a tea.Cmd runs
// rotorSettingsFor off the Bubble Tea goroutine) while the Update goroutine
// writes it on APPLY. The field on Config is the persisted map; these helper
// functions guard it with a mutex so concurrent reads/writes never trigger a
// "concurrent map read and map write" panic.
func (c *Config) GetRadioSettings(stationStr string) (RadioSettings, bool) {
	radioMu.RLock()
	defer radioMu.RUnlock()
	if c.RadioSettings == nil {
		return RadioSettings{}, false
	}
	s, ok := c.RadioSettings[stationStr]
	return s, ok
}

func (c *Config) SetRadioSettings(stationStr string, s RadioSettings) {
	radioMu.Lock()
	defer radioMu.Unlock()
	if c.RadioSettings == nil {
		c.RadioSettings = make(map[string]RadioSettings)
	}
	c.RadioSettings[stationStr] = s
}

func (c *Config) DeleteRadioSettings(stationStr string) {
	radioMu.Lock()
	defer radioMu.Unlock()
	if c.RadioSettings == nil {
		return
	}
	delete(c.RadioSettings, stationStr)
}

// RadioSettingsSnapshot returns a copy of the station-settings map safe for
// use on the caller's goroutine (e.g. by code that needs to enumerate keys).
func (c *Config) RadioSettingsSnapshot() map[string]RadioSettings {
	radioMu.RLock()
	defer radioMu.RUnlock()
	out := make(map[string]RadioSettings, len(c.RadioSettings))
	for k, v := range c.RadioSettings {
		out[k] = v
	}
	return out
}
