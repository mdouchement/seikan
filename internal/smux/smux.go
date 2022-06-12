package smux

import (
	"fmt"
	"regexp"
)

// A Tunnel holds details about the bidirectional multiplexed streaming tunnel.
type Tunnel struct {
	Source       string
	Remote       string
	Destination  string
	IgnoreErrors []*regexp.Regexp
}

func (t Tunnel) String() string {
	return fmt.Sprintf("%s <-> %s <-> %s", t.Source, t.Remote, t.Destination)
}
