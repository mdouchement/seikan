package control

import (
	"encoding/binary"
	"io"
	"net"

	"github.com/fxamacker/cbor/v2"
	"github.com/pkg/errors"
)

// A PDU is the data sent and received through a net connection for control purposes.
type PDU interface {
	RawHeader() *Header
	Size() int
	Version() int
	ControlID() ID
	PID() string
}

// Do sends the given pdu and returns the pdu response.
func Do(c net.Conn, pdu PDU) (PDU, error) {
	err := EncodeTo(c, pdu)
	if err != nil {
		return nil, errors.Wrap(err, "encode")
	}

	pid := pdu.PID()
	cid := pdu.ControlID()

	pdu, err = Decode(c)
	if err != nil {
		return nil, errors.Wrap(err, "decode")
	}

	if pdu.PID() != pid && (pdu.ControlID() != respIDFor(cid) || pdu.ControlID() != ErrorID) {
		return nil, errors.Errorf("invalid %s response", cid)
	}

	if pdu.ControlID() == ErrorID {
		return nil, pdu.(error)
	}

	return pdu, nil
}

// Encode encodes PDU to binary data.
func Encode(pdu PDU) ([]byte, error) {
	payload, err := cbor.Marshal(pdu)
	if err != nil {
		return nil, errors.Wrap(err, "CBOR payload")
	}

	hdr := pdu.RawHeader()
	hdr.size = uint16(hdr.HeaderSize() + len(payload))

	p := make([]byte, int(hdr.size))
	binary.BigEndian.PutUint16(p[:2], hdr.size)
	p[2] = hdr.version
	p[3] = byte(hdr.cid)

	l := len(hdr.pid)
	for i := 0; i < l; i++ {
		p[4+i] = hdr.pid[i]
	}
	p[4+l] = 0x00

	copy(p[4+l+1:], payload)

	return p, nil
}

// EncodeTo encodes PDU to binary data in w.
func EncodeTo(w io.Writer, pdu PDU) error {
	p, err := Encode(pdu)
	if err != nil {
		return err
	}
	_, err = w.Write(p)
	return err
}

// Decode decodes binary PDU data.
func Decode(r io.Reader) (PDU, error) {
	hdr, err := DecodeHeader(r)
	if err != nil {
		return nil, errors.Wrap(err, "header")
	}

	p := make([]byte, hdr.Size()-hdr.HeaderSize())
	_, err = io.ReadFull(r, p)
	if err != nil {
		return nil, errors.Wrap(err, "CBOR payload")
	}

	var pdu PDU

	switch hdr.cid {
	case ErrorID:
		pdu = &Error{
			Header: hdr,
		}
	case InboundsID:
		pdu = &Inbounds{
			Header: hdr,
		}
	case InboundsRespID:
		pdu = &InboundsResp{
			Header: hdr,
		}
	case BindCSID:
		pdu = &BindCS{
			Header: hdr,
		}
	case BindCSRespID:
		pdu = &BindCSResp{
			Header: hdr,
		}
	case BindSCID:
		pdu = &BindSC{
			Header: hdr,
		}
	case BindSCRespID:
		pdu = &BindSCResp{
			Header: hdr,
		}
	}

	if err := cbor.Unmarshal(p, pdu); err != nil {
		return nil, errors.Wrap(err, "CBOR payload")
	}

	return pdu, nil
}

func respIDFor(id ID) ID {
	switch id {
	case InboundsID:
		return InboundsRespID
	case BindCSID:
		return BindCSRespID
	case BindSCID:
		return BindSCRespID
	default:
		return ID(0x00)
	}
}
