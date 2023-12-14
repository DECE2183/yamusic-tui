package api

import (
	"errors"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	minBufferSize    = 256
	bufferFrameScale = 3
	readTimeout      = 100 * time.Millisecond
)

var errOutOfSize = errors.New("position is out of data size")

type HttpReadSeeker struct {
	source       io.ReadCloser
	bufferTimer  *time.Ticker
	readHappened chan struct{}
	readBuffer   []byte
	readIndex    int64
	totalSize    int64
	done         bool
	mux          sync.Mutex
}

func newReadSeaker(rc io.ReadCloser, totalSize int64) *HttpReadSeeker {
	rs := HttpReadSeeker{
		source:       rc,
		totalSize:    totalSize,
		bufferTimer:  time.NewTicker(readTimeout),
		readHappened: make(chan struct{}),
	}
	return &rs
}

func (h *HttpReadSeeker) bufferNextFrame(size int64) {
	if h.totalSize == int64(len(h.readBuffer)) {
		return
	}

	if size < minBufferSize {
		size = minBufferSize
	}

	for {
		h.mux.Lock()

		buf := make([]byte, size)
		n, err := h.source.Read(buf)
		if err == nil || err == io.EOF {
			h.readBuffer = append(h.readBuffer, buf[:n]...)
			if err == io.EOF {
				h.mux.Unlock()
				return
			}
		}

		h.mux.Unlock()

		// await next Read call or timer expiration
		select {
		case <-h.bufferTimer.C:
			size += minBufferSize
			if size > h.totalSize-int64(len(h.readBuffer)) {
				size = h.totalSize - int64(len(h.readBuffer))
			}
			continue
		case <-h.readHappened:
			return
		}
	}
}

func (h *HttpReadSeeker) Close() error {
	var err error

	h.mux.Lock()
	defer h.mux.Unlock()

	h.readBuffer = nil
	if h.totalSize > int64(len(h.readBuffer)) {
		h.bufferTimer.Stop()
		close(h.readHappened)
	}

	if h.source != nil {
		err = h.source.Close()
	}

	return err
}

func (h *HttpReadSeeker) Length() int64 {
	return int64(h.totalSize)
}

func (h *HttpReadSeeker) Read(dest []byte) (n int, err error) {
	h.mux.Lock()

	readBufLen := int64(len(h.readBuffer))
	destLen := int64(len(dest))

	if readBufLen < h.totalSize && h.source != nil {
		// indicate buffering goroutine that Read was called
		h.bufferTimer.Stop()
		close(h.readHappened)
	}

	if h.readIndex >= readBufLen {
		newFrame := make([]byte, (h.readIndex-readBufLen)+destLen)
		n, err = io.ReadFull(h.source, newFrame)
		h.readBuffer = append(h.readBuffer, newFrame[:n]...)
		if h.readIndex < int64(len(h.readBuffer)) {
			copy(dest, h.readBuffer[h.readIndex:])
		} else {
			err = io.EOF
		}
		h.readIndex += int64(n) - (h.readIndex - readBufLen)
	} else {
		var unbufferedLen int

		endIndex := h.readIndex + destLen
		if endIndex > readBufLen {
			endIndex = readBufLen
		}
		bufferedPart := h.readBuffer[h.readIndex:endIndex]

		if destLen-int64(len(bufferedPart)) > 0 {
			unbufferedPart := make([]byte, destLen-int64(len(bufferedPart)))
			unbufferedLen, err = h.source.Read(unbufferedPart)
			unbufferedPart = unbufferedPart[:unbufferedLen]
			copy(dest, append(bufferedPart, unbufferedPart...))
			n = len(bufferedPart) + unbufferedLen
			h.readBuffer = append(h.readBuffer, unbufferedPart...)
		} else {
			copy(dest, bufferedPart)
			n = len(bufferedPart)
		}

		h.readIndex += int64(n)
		if h.readIndex >= h.totalSize {
			err = io.EOF
		}
	}

	if err != nil {
		if err == io.EOF {
			h.source.Close()
			h.source = nil
			h.done = true
		} else if err == http.ErrBodyReadAfterClose {
			err = io.EOF
		}
	} else if h.totalSize > int64(len(h.readBuffer)) {
		h.readHappened = make(chan struct{})
		h.bufferTimer.Reset(readTimeout)
		go h.bufferNextFrame(destLen * bufferFrameScale)
	}

	h.mux.Unlock()
	return
}

func (h *HttpReadSeeker) Seek(offset int64, whence int) (pos int64, err error) {
	h.mux.Lock()

	switch whence {
	case io.SeekStart:
		pos = offset
	case io.SeekCurrent:
		pos = h.readIndex + offset
	case io.SeekEnd:
		pos = h.totalSize + offset
	}

	if pos < 0 || pos > h.totalSize {
		pos = h.readIndex
		err = errOutOfSize
	} else {
		if pos == h.totalSize {
			h.done = true
		} else {
			h.done = false
		}
		h.readIndex = pos
	}

	h.mux.Unlock()
	return
}

func (h *HttpReadSeeker) IsDone() bool {
	return h.done
}

func (h *HttpReadSeeker) Progress() float64 {
	return float64(h.readIndex) / float64(h.totalSize)
}

func (h *HttpReadSeeker) BufferingProgress() float64 {
	return float64(len(h.readBuffer)) / float64(h.totalSize)
}
