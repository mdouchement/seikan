package noise

import (
	"io"
	"net"
)

// A ChunkedConn is able to encrypt/decrypt a net Conn by using chunked encryption.
type ChunkedConn struct {
	net.Conn
	cs io.ReadWriter
}

// NewChunkedConn returns a noise chunked stream.
// Given stream is encrypted by chunks of 0xFFFF max size.
func NewChunkedConn(c net.Conn, cipher Cipher) net.Conn {
	return &ChunkedConn{
		Conn: c,
		cs:   NewChunkedStream(c, cipher),
	}
}

func (c *ChunkedConn) Read(p []byte) (n int, err error) {
	return c.cs.Read(p)
}

func (c *ChunkedConn) Write(p []byte) (n int, err error) {
	return c.cs.Write(p)
}
