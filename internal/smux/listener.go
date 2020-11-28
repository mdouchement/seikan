package smux

import (
	"net"

	"github.com/mdouchement/logger"
)

// A DropListener implements a net.Listener and drops any connection that vannot be accepted.
type DropListener struct {
	net.Listener
	address string
	log     logger.Logger
	connCh  chan net.Conn
}

// NewDropListener returns a new DropListener.
func NewDropListener(log logger.Logger, address string, l net.Listener) *DropListener {
	li := &DropListener{
		log:      log,
		address:  address,
		Listener: l,
		connCh:   make(chan net.Conn, 1),
	}

	go li.serve()
	return li
}

// Accept implements net.Listener.
func (l *DropListener) Accept() <-chan net.Conn {
	return l.connCh
}

// Address implements net.Listener.
func (l *DropListener) Address() string {
	return l.address
}

func (l *DropListener) serve() {
	for {
		conn, err := l.Listener.Accept()
		if err != nil {
			l.log.WithError(err).Warnf("failed to listening on %q: %s", l.Listener.Addr(), err)
			continue
		}

		select {
		case l.connCh <- conn:
		default:
			conn.Close()
			l.log.Infof("No session for %s", conn.LocalAddr())
		}
	}
}
