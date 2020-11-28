package control

import (
	"bytes"
	"fmt"
	"io"
)

// ID is the identifier of the control PDU.
type ID uint8

// Control identifier list.
const (
	ErrorID        ID = 0x01
	InboundsID     ID = 0x02
	InboundsRespID ID = 0x03
	BindCSID       ID = 0x04
	BindCSRespID   ID = 0x05
	BindSCID       ID = 0x06
	BindSCRespID   ID = 0x07
)

func (id ID) String() string {
	switch id {
	case ErrorID:
		return "error"
	case InboundsID:
		return "inbounds"
	case InboundsRespID:
		return "inbounds_resp"
	case BindCSID:
		return "bind_cs"
	case BindCSRespID:
		return "bind_cs_resp"
	case BindSCID:
		return "bind_sc"
	case BindSCRespID:
		return "bind_sc_resp"
	default:
		return fmt.Sprintf("%X", uint8(id))
	}
}

func readNulTerminatedString(r io.Reader) (string, error) {
	buf := bytes.NewBuffer(nil)
	p := make([]byte, 1)

	for {
		if _, err := r.Read(p); err != nil {
			return "", err
		}

		if p[0] != 0x00 {
			buf.Write(p)
		} else {
			return buf.String(), nil
		}
	}
}
