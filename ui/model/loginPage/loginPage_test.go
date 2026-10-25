package loginpage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/config"

	tea "github.com/charmbracelet/bubbletea"
)

func newTestLogin(t *testing.T) *Model {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	if err := config.InitialLoad(); err != nil {
		t.Fatal(err)
	}
	config.Current.DeviceLogin = true
	m := New()
	m.resize(100, 30)
	return m
}

func testCode() *api.DeviceCode {
	return &api.DeviceCode{DeviceCode: "dc", UserCode: "uc", Interval: 5, ExpiresIn: 300}
}

func TestStaleDeviceCodeIsDropped(t *testing.T) {
	m := newTestLogin(t)
	m.codeGen = 2
	m.requesting = true

	m.Update(deviceCodeMsg{gen: 1, code: testCode()})

	if m.code != nil {
		t.Fatalf("stale response replaced the current code: %+v", m.code)
	}
	if !m.requesting {
		t.Fatal("the current request is no longer in flight")
	}
}

func TestTransientPollErrorRetries(t *testing.T) {
	m := newTestLogin(t)
	m.codeGen = 1
	m.code = testCode()

	_, cmd := m.Update(devicePollMsg{gen: 1, err: context.DeadlineExceeded})

	if cmd == nil {
		t.Fatal("a transient error must be retried")
	}
	if m.status != "" {
		t.Fatalf("unexpected status: %q", m.status)
	}
}

func TestTerminalPollErrorStops(t *testing.T) {
	m := newTestLogin(t)
	m.codeGen = 1
	m.code = testCode()

	_, cmd := m.Update(devicePollMsg{gen: 1, err: &api.HTTPStatusError{Status: 400}})

	if cmd != nil {
		t.Fatal("a terminal error must stop polling")
	}
	if !strings.Contains(m.status, "Browser login failed") {
		t.Fatalf("unexpected status: %q", m.status)
	}
}

func TestSaveFailureCanBeRetried(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", config.DirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0755) })
	if err := config.InitialLoad(); err != nil {
		t.Fatal(err)
	}
	config.Current.DeviceLogin = true

	m := New()
	m.codeGen = 1
	m.code = testCode()

	_, cmd := m.Update(devicePollMsg{gen: 1, token: "token"})
	if cmd != nil || !m.saveFailed {
		t.Fatalf("save failure must not quit: cmd=%v saveFailed=%v", cmd, m.saveFailed)
	}

	if _, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter}); cmd != nil {
		t.Fatal("retry with the directory still read-only must not quit")
	}

	if err := os.Chmod(dir, 0755); err != nil {
		t.Fatal(err)
	}
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("a successful retry must quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.Quit, got %T", cmd())
	}
}
