package api

import (
	"context"
	"io"
	"time"
)

type TimeLimitedReader struct {
	timer   *time.Timer
	timeout time.Duration
	cancel  context.CancelFunc
	ctx     context.Context
	body    io.ReadCloser
}

func NewTimeLimitedReader(r io.ReadCloser, ctx context.Context, ctxCancel context.CancelFunc, timeout time.Duration) *TimeLimitedReader {
	tr := &TimeLimitedReader{
		timer:   time.NewTimer(timeout),
		timeout: timeout,
		cancel:  ctxCancel,
		ctx:     ctx,
		body:    r,
	}

	go tr.deadline()
	return tr
}

func (r *TimeLimitedReader) Read(dest []byte) (int, error) {
	r.timer.Reset(r.timeout)
	return r.body.Read(dest)
}

func (r *TimeLimitedReader) Close() error {
	r.cancel()
	r.timer.Stop()
	return r.body.Close()
}

func (r *TimeLimitedReader) deadline() {
	for {
		select {
		case <-r.timer.C:
			r.cancel()
			return
		case <-r.ctx.Done():
			return
		}
	}
}
