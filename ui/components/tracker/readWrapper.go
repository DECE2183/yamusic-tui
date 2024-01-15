package tracker

import (
	"io"
	"time"

	"github.com/dece2183/yamusic-tui/api"

	tea "github.com/charmbracelet/bubbletea"
	mp3 "github.com/dece2183/go-stream-mp3"
)

type readWrapper struct {
	program         *tea.Program
	decoder         *mp3.Decoder
	trackReader     *api.HttpReadSeeker
	trackDurationMs int
	lastUpdateTime  time.Time
	trackStartTime  time.Time
}

func (w *readWrapper) Read(dest []byte) (n int, err error) {
	if w.trackReader == nil {
		err = io.EOF
		return
	}

	n, err = w.decoder.Read(dest)
	if err != nil && err != io.EOF {
		w.trackReader.Close()
		w.trackReader = nil
		go w.program.Send(STOP)
		return
	}

	if w.trackReader.IsDone() {
		w.trackReader.Close()
		w.trackReader = nil
		go w.program.Send(NEXT)
	} else if time.Since(w.lastUpdateTime) > time.Millisecond*33 {
		w.lastUpdateTime = time.Now()
		fraction := ProgressControl(w.trackReader.Progress())
		go w.program.Send(fraction)
	}

	return
}

func (w *readWrapper) Seek(offset int64, whence int) (int64, error) {
	return w.decoder.Seek(offset, whence)
}
