package client

import (
	"github.com/mdouchement/basex"
	"github.com/mdouchement/logger"
	"github.com/mdouchement/seikan/internal/config"
	"github.com/mdouchement/seikan/internal/control"
	"github.com/mdouchement/seikan/internal/seikan"
	"github.com/mdouchement/seikan/internal/smux"
	"github.com/pkg/errors"
)

// Inbound handles server to client tunneling.
type Inbound struct {
	log          logger.Logger
	cfg          config.Client
	destinations map[string]bool
}

// NewInbound returns a new Inbound.
func NewInbound(cfg config.Client, l logger.Logger) (in *Inbound, err error) {
	in = &Inbound{
		cfg: cfg,
		log: l,
	}

	in.destinations, err = in.getDestinations()
	return in, err
}

// Establish establishes the client's tunnels.
// It retries in case of error.
func (in *Inbound) Establish() {
	for destination := range in.destinations {
		go in.retryableEstablish(destination)
	}
}

func (in *Inbound) retryableEstablish(destination string) {
	seikan.Retry(func(prev error) error {
		err := in.establish(destination)
		if seikan.IsRetryNewError(prev, err) {
			in.log.Errorf("closed (%s)", err)
			return err
		}
		in.log.Debug("closed (%s)", err)
		return err
	})
}

func (in *Inbound) establish(destination string) error {
	log := in.log.WithPrefixf("[%s]", basex.GenerateID()).WithPrefix("[ingoing ]")

	tun := smux.Tunnel{
		Source:      "remote_side",
		Remote:      in.cfg.Server.Address,
		Destination: destination,
	}

	c, err := connect(log, in.cfg, tun)
	if err != nil {
		return err
	}
	defer c.Close()

	//

	log.Info("Performing bind_sc control")
	bind := control.NewBindSC()
	bind.Identifier = in.cfg.Identifier
	bind.Address = tun.Destination

	_, err = control.Do(c, bind)
	if err != nil {
		return errors.Wrap(err, "control")
	}

	//

	log.Infof("Accepting destination %s", tun.Destination)

	smux, err := smux.NewServer(log, tun, c)
	if err != nil {
		return errors.Wrap(err, "failed to initialize smux session")
	}
	defer smux.Close()

	return smux.Listen()
}

func (in *Inbound) getDestinations() (map[string]bool, error) {
	log := in.log.WithPrefixf("[%s]", basex.GenerateID()).WithPrefix("[ingoing ]")

	c, err := connect(log, in.cfg, smux.Tunnel{Remote: in.cfg.Server.Address})
	if err != nil {
		return nil, err
	}
	defer c.Close()

	//

	log.Info("Performing inbounds control")
	bind := control.NewInbounds()
	bind.Identifier = in.cfg.Identifier

	resp, err := control.Do(c, bind)
	if err != nil {
		return nil, errors.Wrap(err, "control")
	}

	//

	m := make(map[string]bool)
	for _, wanted := range resp.(*control.InboundsResp).Inbounds {
		for _, allowed := range in.cfg.AllowList {
			if wanted == allowed {
				m[wanted] = true
				break
			}
		}

		if !m[wanted] {
			in.log.Warnf("Dropped destination %s", wanted)
		}
	}

	return m, errors.Wrap(err, "failed to get destinations")
}
