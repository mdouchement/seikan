package control

import "github.com/mdouchement/basex"

// Inbounds is asked by the client to the server to get the server -> client tunnels.
type Inbounds struct {
	*Header    `cbor:"-"`
	Identifier string `cbor:"identifier"`
}

// NewInbounds returns a new Inbounds.
func NewInbounds() *Inbounds {
	return &Inbounds{
		Header: &Header{
			version: 0x01,
			cid:     InboundsID,
			pid:     basex.GenerateID(),
		},
	}
}

// InboundsResp is the response to Inbounds.
type InboundsResp struct {
	*Header  `cbor:"-"`
	Inbounds []string `cbor:"inbounds"`
}

// NewInboundsResp returns a new InboundsResp.
func NewInboundsResp(id string) *InboundsResp {
	return &InboundsResp{
		Header: &Header{
			version: 0x01,
			cid:     InboundsRespID,
			pid:     id,
		},
	}
}
