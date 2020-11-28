package snet

import (
	"net"

	"github.com/hashicorp/yamux"
	"github.com/pkg/errors"
)

func IsTimeout(err error) bool {
	err = errors.Cause(err)

	if err, ok := err.(*net.OpError); ok {
		return err.Timeout() || err.Unwrap() == yamux.ErrTimeout
	}

	if err, ok := err.(net.Error); ok {
		return err.Timeout()
	}

	if err == yamux.ErrTimeout {
		return true
	}
	return false
}
