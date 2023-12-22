package mainpage

import (
	"io"
	"time"
	"yamusic/api"
	"yamusic/ui/model"

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
		go w.program.Send(model.PLAYER_STOP)
		return
	}

	if w.trackReader.IsDone() {
		w.trackReader.Close()
		w.trackReader = nil
		go w.program.Send(model.PLAYER_NEXT)
	} else if time.Since(w.lastUpdateTime) > time.Millisecond*33 {
		w.lastUpdateTime = time.Now()
		fraction := model.ProgressControl(w.trackReader.Progress())
		go w.program.Send(fraction)
	}

	return
}

func (w *readWrapper) Seek(offset int64, whence int) (int64, error) {
	return w.decoder.Seek(offset, whence)
}
