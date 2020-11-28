package snet

import (
	"io"
	"net"

	"github.com/klauspost/compress/zstd"
)

// A CompressConn is able to compress/decompress a net Conn by using zstandard codec.
// It does not close its given net.Conn.
type CompressConn struct {
	net.Conn
	r *zstd.Decoder
	w *zstd.Encoder
}

// Compress returns a CompressConn.
func Compress(c net.Conn) (net.Conn, error) {
	r, err := zstd.NewReader(c, zstd.WithDecoderConcurrency(2), zstd.WithDecoderLowmem(true))
	if err != nil {
		return nil, err
	}

	w, err := zstd.NewWriter(c, zstd.WithEncoderConcurrency(2), zstd.WithWindowSize(128<<10))
	if err != nil {
		return nil, err
	}

	return &CompressConn{
		Conn: c,
		r:    r,
		w:    w,
	}, nil
}

func (c *CompressConn) Read(p []byte) (n int, err error) {
	n, err = c.r.Read(p)
	if err == io.ErrUnexpectedEOF {
		// ErrUnexpectedEOF is returned when connection is closed.
		// FIXME: Can ErrUnexpectedEOF be returned even if the connection is not closed?
		err = io.EOF
	}
	return
}

func (c *CompressConn) Write(p []byte) (n int, err error) {
	defer c.w.Flush()
	return c.w.Write(p)
}

// Close implements io.Close.
func (c *CompressConn) Close() error {
	c.r.Close()
	return c.w.Close()
}
