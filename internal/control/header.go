package control

import (
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

const (
	staticHeaderSize  = 4 // size + version + cid
	minimalHeaderSize = 6 // size + version + cid + pid + NUL
)

// Header is a PDU header.
type Header struct {
	size    uint16
	version uint8
	cid     ID
	pid     string
}

// DecodeHeader decodes binary PDU header data.
func DecodeHeader(r io.Reader) (*Header, error) {
	hdr := &Header{}

	p := make([]byte, staticHeaderSize)
	_, err := io.ReadFull(r, p)
	if err != nil {
		return nil, err
	}

	hdr.size = binary.BigEndian.Uint16(p)
	if hdr.size < minimalHeaderSize {
		return nil, errors.Errorf("PDU too small: %d < %d", hdr.size, minimalHeaderSize)
	}

	hdr.version = p[2]
	hdr.cid = ID(p[3])

	hdr.pid, err = readNulTerminatedString(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not read PDU identifier")
	}

	return hdr, nil
}

// RawHeader returns the Header.
func (h *Header) RawHeader() *Header {
	return h
}

// HeaderSize returns the size of the header.
func (h *Header) HeaderSize() int {
	return staticHeaderSize + len(h.pid) + 1 // NUL char
}

// Size returns the PDU size.
func (h *Header) Size() int {
	return int(h.size)
}

// Version returns the PDU version.
func (h *Header) Version() int {
	return int(h.version)
}

// ControlID returns the PDU control ID.
func (h *Header) ControlID() ID {
	return h.cid
}

// PID returns the PDU ID.
func (h *Header) PID() string {
	return h.pid
}
