package radioconfig

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dece2183/yamusic-tui/config"
	"github.com/dece2183/yamusic-tui/ui/model"
	"github.com/dece2183/yamusic-tui/ui/style"
)

type Control uint

const (
	APPLY Control = iota
	CANCEL
	CURSOR_UP
	CURSOR_DOWN
	VALUE_LEFT
	VALUE_RIGHT
)

// row is one tunable setting shown in the dialog: a name plus the current value.
type row struct {
	name    string
	display string
}

type Model struct {
	width, height int

	rows       []row
	cursor     int
	stationStr string

	Title string
}

// New builds the radio-configuration dialog for the given station (formatted
// StationId "type:tag"). It starts from the persisted settings so the user sees
// the current values; Result() yields the edited set on Apply.
func New(station string) *Model {
	rc := persistedSettings(station)
	return &Model{
		rows: []row{
			{name: "language", display: rc.Language},
			{name: "diversity", display: rc.Diversity},
			{name: "mood", display: fmt.Sprintf("%.2f", rc.Mood)},
			{name: "energy", display: fmt.Sprintf("%.2f", rc.Energy)},
		},
		stationStr: station,
		Title:      "Radio configuration",
	}
}

// persistedSettings returns the settings for this station, falling back to the
// shared "default" entry and then to the server defaults (empty).
func persistedSettings(stationStr string) config.RadioSettings {
	if stationStr != "" {
		if s, ok := config.Current.GetRadioSettings(stationStr); ok {
			return s
		}
	}
	if s, ok := config.Current.GetRadioSettings("default"); ok {
		return s
	}
	return config.RadioSettings{}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) View() string {
	lines := make([]string, 0, len(m.rows))
	for i := range m.rows {
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}
		lines = append(lines, prefix+m.rows[i].name+": "+m.rows[i].display)
	}
	return lipgloss.JoinVertical(
		lipgloss.Left,
		style.DialogTitleStyle.Render(m.Title),
		style.DialogBoxStyle.MaxWidth(m.width).Render(
			lipgloss.JoinVertical(lipgloss.Left, lines...),
		),
	)
}

func (m *Model) Update(message tea.Msg) (*Model, tea.Cmd) {
	controls := config.Current.Controls

	switch msg := message.(type) {
	case tea.KeyMsg:
		keypress := msg.String()

		switch {
		case controls.Apply.Contains(keypress):
			return m, model.Cmd(APPLY)
		case controls.Cancel.Contains(keypress):
			return m, model.Cmd(CANCEL)
		case controls.CursorUp.Contains(keypress):
			if m.cursor > 0 {
				m.cursor--
			}
			return m, model.Cmd(CURSOR_UP)
		case controls.CursorDown.Contains(keypress):
			if m.cursor < len(m.rows)-1 {
				m.cursor++
			}
			return m, model.Cmd(CURSOR_DOWN)
		case keypress == "right":
			m.adjust(+1)
			return m, model.Cmd(VALUE_RIGHT)
		case keypress == "left":
			m.adjust(-1)
			return m, model.Cmd(VALUE_LEFT)
		}
	}

	return m, nil
}

// adjust changes the currently selected row. Language and diversity cycle
// through a fixed set; mood and energy step by 0.1 within [0, 1].
func (m *Model) adjust(delta int) {
	r := &m.rows[m.cursor]
	switch r.name {
	case "language":
		r.display = cycle([]string{"", "any", "not-russian", "russian"}, r.display, delta)
	case "diversity":
		r.display = cycle([]string{"", "default", "moderate", "high", "off"}, r.display, delta)
	case "mood", "energy":
		v := parseFloat32(r.display) + float32(delta)*0.1
		if v < 0 {
			v = 0
		}
		if v > 1 {
			v = 1
		}
		r.display = fmt.Sprintf("%.2f", v)
	}
}

// Result assembles the edited settings into a RadioSettings value.
func (m *Model) Result() config.RadioSettings {
	rc := persistedSettings(m.stationStr)
	rc.Language = m.rows[0].display
	rc.Diversity = m.rows[1].display
	rc.Mood = parseFloat32(m.rows[2].display)
	rc.Energy = parseFloat32(m.rows[3].display)
	return rc
}

// Station returns the configured station id ("type:tag").
func (m *Model) Station() string {
	return m.stationStr
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// cycle moves current through values by delta, wrapping around. An empty value
// is treated as index 0, so an unset field starts the cycle from the first
// non-default option.
func cycle(values []string, current string, delta int) string {
	idx := 0
	for i := range values {
		if values[i] == current {
			idx = i
			break
		}
	}
	idx = (idx + delta) % len(values)
	if idx < 0 {
		idx += len(values)
	}
	return values[idx]
}

func parseFloat32(s string) float32 {
	var v float32
	_, _ = fmt.Sscanf(s, "%f", &v)
	return v
}
