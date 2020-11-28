package control

import "github.com/mdouchement/basex"

// BindSC is used by the client to open a tunnel.
// This tunnel will accept a bidirectional connection from server to client (the server start the listener).
type BindSC struct {
	*Header    `cbor:"-"`
	Identifier string `cbor:"identifier"`
	Address    string `cbor:"address"`
}

// NewBindSC returns a new BindSC.
func NewBindSC() *BindSC {
	return &BindSC{
		Header: &Header{
			version: 0x01,
			cid:     BindSCID,
			pid:     basex.GenerateID(),
		},
	}
}

// BindSCResp is the response to BindSC.
type BindSCResp struct {
	*Header `cbor:"-"`
}

// NewBindSCResp returns a new BindSCResp.
func NewBindSCResp(id string) *BindSCResp {
	return &BindSCResp{
		Header: &Header{
			version: 0x01,
			cid:     BindSCRespID,
			pid:     id,
		},
	}
}
