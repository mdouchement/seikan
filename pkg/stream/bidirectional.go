package stream

import "io"

// Bidirectional is a bidirectional stream.
type Bidirectional struct {
	C1 *Endpoint
	C2 *Endpoint
}

// NewBidirectional returns a new Bidirectional stream.
func NewBidirectional() *Bidirectional {
	c1r, c2w := io.Pipe()
	c2r, c1w := io.Pipe()

	return &Bidirectional{
		C1: &Endpoint{
			r: c1r,
			w: c1w,
		},
		C2: &Endpoint{
			r: c2r,
			w: c2w,
		},
	}
}

// Close implements io.Closer.
func (s *Bidirectional) Close() error {
	if err := s.C1.Close(); err != nil {
		return err
	}
	if err := s.C2.Close(); err != nil {
		return err
	}
	return nil
}
