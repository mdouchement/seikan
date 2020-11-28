package snet

import (
	"net"
)

// CustomConnCloser returns a net.Conn from c and the close policy handler.
func CustomConnCloser(c net.Conn, close func() error) net.Conn {
	return &conncloser{
		Conn:  c,
		close: close,
	}
}

// NopConnCloser returns a Conn with a no-op Close method wrapping
// the provided Conn c.
func NopConnCloser(c net.Conn) net.Conn {
	return &conncloser{
		Conn: c,
		close: func() error {
			return nil
		},
	}
}

type conncloser struct {
	net.Conn
	close func() error
}

func (c *conncloser) Close() error {
	return c.close()
}
