package server

import (
	"net"

	"github.com/mdouchement/logger"
	"github.com/mdouchement/seikan/internal/config"
	"github.com/mdouchement/seikan/internal/seikan"
	"github.com/mdouchement/seikan/internal/smux"
	"github.com/pkg/errors"
)

// Outbound handles server to client tunneling.
type Outbound struct {
	log       logger.Logger
	cfg       config.Server
	listeners map[string]*smux.DropListener
}

// NewOutbound returns a new Outbound.
func NewOutbound(cfg config.Server, log logger.Logger) (out *Outbound, err error) {
	out = &Outbound{
		cfg:       cfg,
		log:       log.WithPrefix("[outgoing]"),
		listeners: make(map[string]*smux.DropListener, len(cfg.Outbounds)),
	}

	for _, o := range cfg.Outbounds {
		l, err := net.Listen("tcp", o.Source)
		if err != nil {
			return nil, err
		}
		key := out.key(o)
		log.Infof("%s listening on %s", key, o.Source)
		out.listeners[key] = smux.NewDropListener(log, o.Source, l)
	}

	return out, err
}

// Establish establishes the tunnel on the given remote connection rc for the given outbound config.
func (out *Outbound) Establish(log logger.Logger, outbound config.Outbound, rc net.Conn) error {
	var found bool
	for _, o := range out.cfg.Outbounds {
		if o.Identifier == outbound.Identifier && o.Destination == outbound.Destination {
			found = true
		}
	}

	if !found {
		return errors.New("outbound configuration not found")
	}

	l, ok := out.listeners[out.key(outbound)]
	if !ok {
		return errors.New("unregistred listener for outbound") // Should never occurs
	}

	tun := smux.Tunnel{
		Source:      l.Address(),
		Remote:      out.cfg.Address,
		Destination: outbound.Destination,
	}

	return out.establish(log, tun, l, rc)
}

func (out *Outbound) establish(log logger.Logger, tun smux.Tunnel, l *smux.DropListener, rc net.Conn) error {
	smux, err := smux.NewClient(log, tun, l, rc)
	if err != nil {
		return errors.Wrap(err, "failed to initialize smux session")
	}
	defer smux.Close()

	return smux.Establish()
}

func (out *Outbound) key(o config.Outbound) string {
	return seikan.CraftKey(o.Identifier, o.Destination)
}
