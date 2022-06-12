package smux

import (
	"fmt"
	"net"
	"regexp"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/mdouchement/logger"
	"github.com/mdouchement/seikan/internal/snet"
	"github.com/pkg/errors"
)

// A Client allows bidirectional multiplexed streaming over an established tunnel.
// It accepts requests on a listener and forwards it to the server.
type Client struct {
	log      logger.Logger
	listener *DropListener
	rc       net.Conn
	session  *yamux.Session
	ignore   []*regexp.Regexp
}

// NewClient returns a new Client.
func NewClient(l logger.Logger, tun Tunnel, listener *DropListener, rc net.Conn) (*Client, error) {
	cfg := yamux.DefaultConfig()
	cfg.EnableKeepAlive = false
	session, err := yamux.Client(snet.NopConnCloser(rc), cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to establish session")
	}

	client := &Client{
		log:      l.WithPrefixf("[smux][%s]", tun.Source),
		listener: listener,
		rc:       rc,
		session:  session,
		ignore:   tun.IgnoreErrors,
	}
	client.log.Infof("Session oppened %s", tun)

	// go client.ping() // FIXME: seems useless if client/server connection is stoped
	return client, nil
}

// Establish establishes multiplexed streaming with the server.
func (cl *Client) Establish() error {
	for {
		select {
		case <-cl.session.CloseChan():
			cl.log.Info("Session closed")
			return nil
		case c := <-cl.listener.Accept():
			go func() {
				defer c.Close()

				stream, err := cl.session.OpenStream()
				if err != nil {
					for _, re := range cl.ignore {
						if re.MatchString(err.Error()) {
							cl.log.WithError(err).Debug("failed to open stream")
							return
						}
					}

					cl.log.WithError(err).Warn("failed to open stream")
					return
				}
				defer stream.Close()

				pipe, err := snet.NewPipe(c, stream)
				if err != nil {
					for _, re := range cl.ignore {
						if re.MatchString(err.Error()) {
							cl.log.WithError(err).Debug("failed to establish pipe")
							return
						}
					}

					cl.log.WithError(err).Warn("failed to establish pipe")
					return
				}
				defer pipe.Close()

				err = pipe.Relay()
				if err != nil && !snet.IsTimeout(err) {
					entry := cl.log.WithFields(logger.M{
						"local":  fmt.Sprintf("%s/%s", pipe.LocalConn().LocalAddr(), pipe.LocalConn().RemoteAddr()),
						"remote": fmt.Sprintf("%s/%s", pipe.RemoteConn().LocalAddr(), pipe.RemoteConn().RemoteAddr()),
					})
					if snet.IsTimeout(err) {
						entry.Debugf("pipe failure (%s)", err)
						return
					}

					for _, re := range cl.ignore {
						if re.MatchString(err.Error()) {
							entry.Debugf("pipe failure (%s)", err)
							return
						}
					}

					entry.Errorf("pipe failure (%s)", err)
				}
			}()
		}
	}
}

func (cl *Client) ping() {
	for {
		_, err := cl.session.Ping()
		if err != nil {
			cl.session.Close()
			return
		}
		time.Sleep(time.Second)
	}
}

// Close implements io.Close.
func (cl *Client) Close() {
	cl.session.Close() // Closes also the given conn to Yamux. Use snet.NopConnCloser to avoid it.
}
