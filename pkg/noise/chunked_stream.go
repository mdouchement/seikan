package noise

import (
	"encoding/binary"
	"errors"
	"io"
)

// nrekey is the n chunks before rekeying the stream.
// It must be a power of 2 minus 1 because we use bitwise-and to compute modulus for more performance.
const nrekey = 16<<10 - 1

// A ChunkedStream is able to encrypt/decrypt a stream by using chunked encryption.
type ChunkedStream struct {
	stream  io.ReadWriter
	cipher  Cipher
	maxsize int
	bufw    []byte
	iw      uint64
	bufr    []byte
	ir      uint64
	chunk   []byte
}

// NewChunkedStream returns a noise chunked stream.
// Given stream is encrypted by chunks of 0xFFFF max size.
func NewChunkedStream(stream io.ReadWriter, cipher Cipher) io.ReadWriter {
	return &ChunkedStream{
		stream:  stream,
		cipher:  cipher,
		maxsize: ChunkSize - cipher.Overhead(),
		bufw:    make([]byte, 0, ChunkSize+SizePrefixLength),
		bufr:    make([]byte, 0, ChunkSize),
	}
}

func (c *ChunkedStream) Read(p []byte) (n int, err error) {
	if len(c.chunk) == 0 {
		if err = c.read(); err != nil {
			return 0, err
		}
	}

	l := len(p)
	lc := len(c.chunk)

	if lc < l {
		copy(p, c.chunk)
		c.chunk = c.chunk[:0]
		return lc, nil
	}

	copy(p, c.chunk[:l])
	c.chunk = c.chunk[l:]
	return l, err
}

// reads chunk by reuing the same underlying array (no extra allocation).
func (c *ChunkedStream) read() (err error) {
	c.ir++
	if c.ir&nrekey == 0 {
		c.cipher.DecryptRekey()
	}

	//

	c.chunk = c.bufr[:]

	_, err = c.stream.Read(c.chunk[:SizePrefixLength])
	if err != nil {
		return err
	}
	n := binary.BigEndian.Uint16(c.chunk[:SizePrefixLength])

	var offset uint16
	for offset < n {
		rn, err := c.stream.Read(c.chunk[offset:n])
		if err != nil {
			return err
		}

		offset += uint16(rn)
	}
	if offset != n {
		return errors.New("invalid read length")
	}

	chunk, err := c.cipher.Decrypt(c.chunk[:0], nil, c.chunk[:n])
	if err != nil {
		return err
	}
	c.chunk = c.chunk[:len(chunk)]
	return nil
}

func (c *ChunkedStream) Write(p []byte) (n int, err error) {
	remaining := len(p)
	var limit int
	var nn int

	// Split p in chunks
	for remaining > 0 {
		limit = remaining
		if limit > c.maxsize {
			limit = c.maxsize
		}

		nn, err = c.write(p[n : n+limit])
		if err != nil {
			return n, err
		}

		remaining -= nn
		n += nn
	}

	return n, nil
}

// write a chunk
func (c *ChunkedStream) write(p []byte) (n int, err error) {
	c.iw++
	if c.iw&nrekey == 0 {
		c.cipher.EncryptRekey()
	}

	//

	n = len(p)
	chunk := c.bufw[:SizePrefixLength]

	p, err = c.cipher.Encrypt(chunk[SizePrefixLength:], nil, p)
	if err != nil {
		return 0, err
	}
	cn := len(p)

	binary.BigEndian.PutUint16(chunk[:SizePrefixLength], uint16(cn)) // Set chunk size
	_, err = c.stream.Write(chunk[:SizePrefixLength+cn])
	return n, err
}
