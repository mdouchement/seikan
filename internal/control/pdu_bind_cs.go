package control

import "github.com/mdouchement/basex"

// BindCS is used by the client to open a tunnel.
// This tunnel will accept a bidirectional connection from client to server (the client start the listener).
type BindCS struct {
	*Header    `cbor:"-"`
	Identifier string `cbor:"identifier"`
	Address    string `cbor:"address"`
}

// NewBindCS returns a new BindCS.
func NewBindCS() *BindCS {
	return &BindCS{
		Header: &Header{
			version: 0x01,
			cid:     BindCSID,
			pid:     basex.GenerateID(),
		},
	}
}

// BindCSResp is the response to BindCS.
type BindCSResp struct {
	*Header `cbor:"-"`
}

// NewBindCSResp returns a new BindCSResp.
func NewBindCSResp(id string) *BindCSResp {
	return &BindCSResp{
		Header: &Header{
			version: 0x01,
			cid:     BindCSRespID,
			pid:     id,
		},
	}
}
