package client

import (
	"fmt"
	"net"

	"github.com/mdouchement/basex"
	"github.com/mdouchement/logger"
	"github.com/mdouchement/seikan/internal/config"
	"github.com/mdouchement/seikan/internal/control"
	"github.com/mdouchement/seikan/internal/seikan"
	"github.com/mdouchement/seikan/internal/smux"
	"github.com/pkg/errors"
)

// Outbound handles client to server tunneling.
type Outbound struct {
	log       logger.Logger
	cfg       config.Client
	listeners map[string]*smux.DropListener
}

// NewOutbound returns a new Outbound.
func NewOutbound(cfg config.Client, l logger.Logger) *Outbound {
	return &Outbound{
		log:       l,
		cfg:       cfg,
		listeners: make(map[string]*smux.DropListener),
	}
}

// Establish establishes the tunnel.
// It retries in case of error.
func (out *Outbound) Establish() error {
	for _, o := range out.cfg.Outbounds {
		l, err := net.Listen("tcp", o.Source)
		if err != nil {
			return err
		}
		out.log.Infof("Listening on %s", o.Source)
		out.listeners[o.Source] = smux.NewDropListener(out.log, o.Source, l)

		//
		//

		go func(o config.Outbound) {
			seikan.Retry(func(prev error) error {
				log := out.log.WithPrefixf("[%s]", basex.GenerateID()).WithPrefix("[outgoing]")
				err := out.establish(log, o.Source, o.Destination)
				if seikan.IsRetryNewError(prev, err) {
					log.Errorf("closed (%s)", err) // TODO: if it's retryable, we should not logs closed?
					return err
				}
				return err
			})
		}(o)
	}

	return nil
}

func (out *Outbound) establish(log logger.Logger, source string, destination string) error {
	tun := smux.Tunnel{
		Source:      source,
		Remote:      out.cfg.Server.Address,
		Destination: destination,
	}

	rc, err := connect(log, out.cfg, tun)
	if err != nil {
		return err
	}
	defer rc.Close()

	//

	log.Info("Performing bind_cs control")
	bind := control.NewBindCS()
	bind.Identifier = out.cfg.Identifier
	bind.Address = tun.Destination

	_, err = control.Do(rc, bind)
	if err != nil {
		return fmt.Errorf("control: %w: %s: %w", seikan.ErrNotRetayable, destination, err)
	}

	//

	smux, err := smux.NewClient(log, tun, out.listeners[source], rc)
	if err != nil {
		return errors.Wrap(err, "failed to initialize smux session")
	}
	defer smux.Close()

	return smux.Establish()
}
