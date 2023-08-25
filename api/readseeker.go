package api

import (
	"errors"
	"io"
	"sync"
)

var errOutOfSize = errors.New("position is out of data size")

type HttpReadSeeker struct {
	source     io.ReadCloser
	readBuffer []byte
	readIndex  int
	totalSize  int
	done       bool
	mux        sync.Mutex
}

func newReadSeaker(rc io.ReadCloser, totalSize int) *HttpReadSeeker {
	rs := HttpReadSeeker{
		source:    rc,
		totalSize: totalSize,
	}
	return &rs
}

func (h *HttpReadSeeker) bufferNextFrame(size int) {
	h.mux.Lock()

	if size < 128 {
		size = 128
	}

	buf := make([]byte, size)
	n, err := h.source.Read(buf)
	if err == nil || err == io.EOF {
		h.readBuffer = append(h.readBuffer, buf[:n]...)
	}

	h.mux.Unlock()
}

func (h *HttpReadSeeker) Close() error {
	h.mux.Lock()
	defer h.mux.Unlock()
	return h.source.Close()
}

func (h *HttpReadSeeker) Read(dest []byte) (n int, err error) {
	h.mux.Lock()

	readBufLef := len(h.readBuffer)

	if h.readIndex >= readBufLef {
		n, err = h.source.Read(dest)
		h.readBuffer = append(h.readBuffer, dest[:n]...)
		h.readIndex += n
	} else {
		var unbufferedLen int
		endIndex := h.readIndex + len(dest)
		if endIndex > readBufLef {
			endIndex = readBufLef
		}
		bufferedPart := h.readBuffer[h.readIndex:endIndex]
		if len(dest)-len(bufferedPart) > 0 {
			unbufferedPart := make([]byte, len(dest)-len(bufferedPart))
			unbufferedLen, err = h.source.Read(unbufferedPart)
			unbufferedPart = unbufferedPart[:unbufferedLen]
			copy(dest, append(bufferedPart, unbufferedPart...))
			n = len(bufferedPart) + unbufferedLen
			h.readBuffer = append(h.readBuffer, unbufferedPart...)
		} else {
			copy(dest, bufferedPart)
			n = len(bufferedPart)
		}
		h.readIndex += n
		if h.readIndex >= h.totalSize {
			err = io.EOF
		}
	}

	if err == io.EOF {
		h.source.Close()
		h.done = true
	} else if err == io.ErrClosedPipe {
		err = io.EOF
	} else {
		go h.bufferNextFrame(n)
	}

	h.mux.Unlock()
	return
}

func (h *HttpReadSeeker) SeekPos(offset int, whence int) (pos int, err error) {
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
