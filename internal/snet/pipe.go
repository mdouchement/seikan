package snet

import (
	"net"

	"github.com/pkg/errors"
)

// A Pipe pipes two net.Conn.
type Pipe struct {
	rc net.Conn
	c  net.Conn
}

// NewPipeTCP returns a new Pipe by opening a new TCP connection on remote.
func NewPipeTCP(c net.Conn, remote string) (*Pipe, error) {
	rc, err := net.Dial("tcp", remote)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to remote")
	}
	rc.(*net.TCPConn).SetKeepAlive(true)

	return NewPipe(c, rc)
}

// NewPipe returns a new Pipe between the given connections.
func NewPipe(c, rc net.Conn) (*Pipe, error) {
	return &Pipe{
		rc: rc,
		c:  c,
	}, nil
}

// Relay runs the pipeline.
func (s *Pipe) Relay() error {
	err := Relay(s.c, s.rc)
	return errors.Wrap(err, "pipe-relay")
}

// LocalConn returns the local connection.
func (s *Pipe) LocalConn() net.Conn {
	return s.c
}

// RemoteConn returns the remote connection.
func (s *Pipe) RemoteConn() net.Conn {
	return s.rc
}

// Close closes both local and remote connections.
func (s *Pipe) Close() {
	s.c.Close()
	s.rc.Close()
}
