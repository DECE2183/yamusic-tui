package log

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dece2183/yamusic-tui/config"
)

type Level int

const (
	LVL_PANIC Level = iota
	LVL_ERROR
	LVL_WARNIGN
	LVL_INFO
)

var lvlName = map[Level]string{
	LVL_PANIC:   "PANIC",
	LVL_ERROR:   "ERROR",
	LVL_WARNIGN: "WARN",
	LVL_INFO:    "INFO",
}

var (
	file *os.File
)

func getLogLocation() (string, error) {
	tempDir := filepath.Join(os.TempDir(), config.DirName)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(tempDir, config.DirName+".log"), nil
}

func Location() string {
	if file == nil {
		return ""
	}

	return file.Name()
}

func Start() {
	path, err := getLogLocation()
	if err != nil {
		return
	}

	file, _ = os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
}

func Stop() {
	file.Close()
	file = nil
}

func Print(lvl Level, format string, args ...any) {
	if file == nil {
		return
	}

	t := time.Now()
	format = fmt.Sprintf("[ %s ][(%d) %02d:%02d:%02d] - ", lvlName[lvl], t.Day(), t.Hour(), t.Minute(), t.Second()) + format + "\n"
	fmt.Fprintf(file, format, args...)
	file.Sync()
}
