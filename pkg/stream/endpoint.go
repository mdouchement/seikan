package stream

import "io"

// Endpoint modelizes one side of a bidirectional stream.
type Endpoint struct {
	r *io.PipeReader
	w *io.PipeWriter
}

// Close implements io.Closer.
func (e *Endpoint) Close() error {
	if err := e.w.Close(); err != nil {
		return err
	}
	if err := e.r.Close(); err != nil {
		return err
	}
	return nil
}

func (e *Endpoint) Read(p []byte) (n int, err error) {
	return e.r.Read(p)
}

func (e *Endpoint) Write(p []byte) (n int, err error) {
	return e.w.Write(p)
}
