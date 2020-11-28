package client

import (
	"net"

	"github.com/hashicorp/go-multierror"
	"github.com/mdouchement/logger"
	"github.com/mdouchement/seikan/internal/config"
	"github.com/mdouchement/seikan/internal/noise"
	"github.com/mdouchement/seikan/internal/seikan"
	"github.com/mdouchement/seikan/internal/smux"
	"github.com/mdouchement/seikan/internal/snet"
	"github.com/pkg/errors"
)

type (
	// A Client dials a server for running Seikan's tunnels.
	Client interface {
		Dial() error
	}

	client struct {
		cfg       config.Client
		log       logger.Logger
		listeners map[string]*smux.DropListener
	}
)

// New returns a new Client.
func New(cfg config.Client, l logger.Logger) Client {
	return &client{
		cfg:       cfg,
		log:       l,
		listeners: make(map[string]*smux.DropListener),
	}
}

// Dial establishes a tunnel with the server.
func (client *client) Dial() error {
	// TUNNEL server to client
	if client.cfg.Inbound {
		inbound, err := NewInbound(client.cfg, client.log)
		if err != nil {
			return err
		}

		inbound.Establish()
	}

	// TUNNEL client to server
	if len(client.cfg.Outbounds) > 0 {
		outbound := NewOutbound(client.cfg, client.log)

		err := outbound.Establish()
		if err != nil {
			return err
		}
	}

	return nil
}

func identity(c config.Client) noise.Identity {
	return noise.Identity{
		Secret: c.Secret,
		Public: c.Public,
	}
}

func recipient(c config.Client) string {
	return c.Server.Public
}

func connect(log logger.Logger, cfg config.Client, t smux.Tunnel) (net.Conn, error) {
	c, err := snet.Dial(t.Remote)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to server %s", t.Remote)
	}
	snet.EnableKeepAlive(c)

	//

	log.Info("Handshake")

	log.Debug("Sending derived identifier")
	derived, err := seikan.KDFGenerate(cfg.Identifier)
	if err != nil {
		c.Close()
		return nil, errors.Wrap(err, "failed to generate derived identifier")
	}

	if _, err = c.Write(derived); err != nil {
		c.Close()
		return nil, errors.Wrap(err, "failed to send derived identifier")
	}

	//

	log.Debug("Performing Noise handshake")
	nc, err := noise.Handshake(c, identity(cfg), recipient(cfg), false)
	if err != nil {
		c.Close()
		return nil, err
	}

	cc, err := snet.Compress(nc)

	//

	close := func() error {
		var result error

		err := c.Close()
		if err != nil {
			result = multierror.Append(result, err)
		}

		err = cc.Close()
		if err != nil {
			result = multierror.Append(result, err)
		}

		return result
	}

	return snet.CustomConnCloser(cc, close), err
}
