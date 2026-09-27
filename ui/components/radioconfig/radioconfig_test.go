package radioconfig

import (
	"testing"

	"github.com/dece2183/yamusic-tui/config"
)

func TestCycle(t *testing.T) {
	values := []string{"", "any", "not-russian", "russian"}
	if got := cycle(values, "", 1); got != "any" {
		t.Errorf("cycle(empty,+1) = %q, want any", got)
	}
	if got := cycle(values, "russian", 1); got != "" {
		t.Errorf("cycle(russian+1) = %q, want wrap to empty", got)
	}
	if got := cycle(values, "", -1); got != "russian" {
		t.Errorf("cycle(-1) = %q, want wrap to last", got)
	}
}

func TestAdjustMoodClamps(t *testing.T) {
	m := New("genre:test")
	m.cursor = 2 // focus the mood row
	m.rows[2].display = "0.90"
	m.adjust(+1) // 1.00
	if m.rows[2].display != "1.00" {
		t.Errorf("mood up = %q, want 1.00", m.rows[2].display)
	}
	m.adjust(+1) // clamp at 1.00
	if m.rows[2].display != "1.00" {
		t.Errorf("mood over-clamped = %q, want 1.00", m.rows[2].display)
	}
	m.adjust(-1) // 0.90
	if m.rows[2].display != "0.90" {
		t.Errorf("mood down = %q, want 0.90", m.rows[2].display)
	}
}

func TestResultFromRows(t *testing.T) {
	m := New("genre:test")
	m.rows[0].display = "russian"
	m.rows[1].display = "high"
	m.rows[2].display = "0.50"
	m.rows[3].display = "0.25"
	res := m.Result()
	if res.Language != "russian" || res.Diversity != "high" ||
		res.Mood != 0.5 || res.Energy != 0.25 {
		t.Errorf("Result = %+v, want russian/high/0.5/0.25", res)
	}
}

func TestPersistedSettingsFallback(t *testing.T) {
	config.Current.RadioSettings = map[string]config.RadioSettings{}
	got := persistedSettings("genre:test")
	if !got.IsZero() {
		t.Errorf("empty settings should be zero, got %+v", got)
	}

	config.Current.RadioSettings = map[string]config.RadioSettings{
		"default": {Language: "not-russian", Mood: 0.4},
	}
	got = persistedSettings("genre:other")
	if got.Language != "not-russian" || got.Mood != 0.4 {
		t.Errorf("default fallback = %+v, want not-russian/0.4", got)
	}
	config.Current.RadioSettings = nil
}
