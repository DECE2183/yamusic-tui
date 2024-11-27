package tracker

import (
	"io"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	mp3 "github.com/dece2183/go-stream-mp3"
	"github.com/dece2183/yamusic-tui/stream"
)

type readWrapper struct {
	program         *tea.Program
	decoder         *mp3.Decoder
	trackReader     *stream.BufferedStream
	trackDurationMs int
	lastUpdateTime  time.Time
}

func (w *readWrapper) Read(dest []byte) (n int, err error) {
	if w.trackReader == nil {
		err = io.EOF
		return
	}

	n, err = w.decoder.Read(dest)
	if err != nil && err != io.EOF {
		// bypass mp3 decoding error after rewinding
		err = nil
	}

	if w.trackReader.IsDone() {
		w.decoder.Seek(0, io.SeekStart)
		w.trackReader.Close()
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
