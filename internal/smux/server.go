package smux

import (
	"fmt"
	"io"
	"net"

	"github.com/hashicorp/yamux"
	"github.com/mdouchement/logger"
	"github.com/mdouchement/seikan/internal/snet"
	"github.com/pkg/errors"
)

// A Server allows bidirectional multiplexed streaming over an established tunnel.
// It receives a stream from the client and forwards its specified destination.
type Server struct {
	rc      net.Conn
	tun     Tunnel
	log     logger.Logger
	session *yamux.Session
}

// NewServer returns a new Server.
func NewServer(l logger.Logger, tun Tunnel, rc net.Conn) (*Server, error) {
	l = l.WithPrefixf("[smux][%s]", tun.Destination)

	cfg := yamux.DefaultConfig()
	cfg.EnableKeepAlive = false
	session, err := yamux.Server(snet.NopConnCloser(rc), cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to establish session")
	}
	l.Infof("Session oppened %s", tun)

	return &Server{
		rc:      rc,
		tun:     tun,
		log:     l,
		session: session,
	}, nil
}

// Listen listens for client's multiplexed streaming to forward to the configured destination.
func (s *Server) Listen() error {
	for {
		sc, err := s.session.Accept()
		if err != nil {
			if err == io.EOF {
				s.log.Info("session closed")
				return nil
			}
			return errors.Wrap(err, "smux: server: failed to accept stream")
		}

		go func() {
			defer sc.Close()

			pipe, err := snet.NewPipeTCP(sc, s.tun.Destination) // TODO: allow UDP too
			if err != nil {
				s.log.WithError(err).Error("failed to establish pipe session")
				return
			}
			defer pipe.Close()

			err = pipe.Relay()
			if err != nil && !snet.IsTimeout(err) {
				entry := s.log.WithFields(logger.M{
					"local":  fmt.Sprintf("%s/%s", pipe.LocalConn().LocalAddr(), pipe.LocalConn().RemoteAddr()),
					"remote": fmt.Sprintf("%s/%s", pipe.RemoteConn().LocalAddr(), pipe.RemoteConn().RemoteAddr()),
				})
				if snet.IsTimeout(err) {
					entry.Debugf("pipe failure (%s)", err)
					return
				}
				entry.Errorf("pipe failure (%s)", err)
			}
		}()
	}
}

// Close implements io.Close.
func (s *Server) Close() {
	s.session.Close() // Closes also the given conn to Yamux. Use snet.NopConnCloser to avoid it.
}
