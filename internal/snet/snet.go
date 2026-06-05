package snet

import (
	"errors"
	"net"

	"github.com/hashicorp/yamux"
)

func IsTimeout(err error) bool {
	if operr, ok := errors.AsType[*net.OpError](err); ok {
		return operr.Timeout() || operr.Unwrap() == yamux.ErrTimeout
	}

	if nerr, ok := errors.AsType[net.Error](err); ok {
		return nerr.Timeout()
	}

	return err == yamux.ErrTimeout
}
