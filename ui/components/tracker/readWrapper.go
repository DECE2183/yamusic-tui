package tracker

import (
	"io"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	mp3 "github.com/dece2183/go-stream-mp3"
	"github.com/dece2183/yamusic-tui/log"
	"github.com/dece2183/yamusic-tui/stream"
)

type readWrapper struct {
	program        *tea.Program
	decoder        *mp3.Decoder
	trackBuffer    *stream.BufferedStream
	trackBuffered  bool
	lastUpdateTime time.Time
}

func (w *readWrapper) NewReader(reader *stream.BufferedStream) {
	var err error

	w.trackBuffered = false
	w.trackBuffer = reader
	w.decoder, err = mp3.NewDecoder(w.trackBuffer)
	if err != nil {
		log.Print(log.LVL_ERROR, "failed to create mp3 decoder: %s", err)
		return
	}

	w.lastUpdateTime = time.Now()
}

func (w *readWrapper) Close() {
	if w.decoder != nil {
		w.decoder.Seek(0, io.SeekStart)
	}

	if w.trackBuffer != nil {
		w.trackBuffer.Close()
	}
}

func (w *readWrapper) Read(dest []byte) (n int, err error) {
	if w.trackBuffer == nil {
		err = io.EOF
		return
	}

	n, err = w.decoder.Read(dest)
	if err != nil && err != io.EOF {
		// bypass mp3 decoding error after rewinding
		log.Print(log.LVL_WARNIGN, "mp3 decoding error: %s", err)
		err = nil
	}

	if w.trackBuffer.IsBuffered() && !w.trackBuffered {
		w.trackBuffered = true
		go w.program.Send(BUFFERING_COMPLETE)
	}

	if w.trackBuffer.IsDone() {
		w.decoder.Seek(0, io.SeekStart)
		w.trackBuffer.Close()
		go w.program.Send(NEXT)
	} else if time.Since(w.lastUpdateTime) > time.Millisecond*33 {
		w.lastUpdateTime = time.Now()
		fraction := ProgressControl(w.trackBuffer.Progress())
		go w.program.Send(fraction)
	}

	return
}

func (w *readWrapper) Seek(offset int64, whence int) (int64, error) {
	return w.decoder.Seek(offset, whence)
}

func (w *readWrapper) Length() int64 {
	return w.trackBuffer.Length()
}

func (w *readWrapper) Progress() float64 {
	return w.trackBuffer.Progress()
}
