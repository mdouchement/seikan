package smux

import (
	"fmt"
)

// A Tunnel holds details about the bidirectional multiplexed streaming tunnel.
type Tunnel struct {
	Source      string
	Remote      string
	Destination string
}

func (t Tunnel) String() string {
	return fmt.Sprintf("%s <-> %s <-> %s", t.Source, t.Remote, t.Destination)
}
