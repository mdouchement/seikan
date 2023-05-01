package client

import (
	"context"

	"github.com/mdouchement/basex"
	"github.com/mdouchement/logger"
	"github.com/mdouchement/seikan/internal/config"
	"github.com/mdouchement/seikan/internal/control"
	"github.com/mdouchement/seikan/internal/filter"
	"github.com/mdouchement/seikan/internal/seikan"
	"github.com/mdouchement/seikan/internal/smux"
	"github.com/pkg/errors"
)

// Inbound handles server to client tunneling.
type Inbound struct {
	log          logger.Logger
	cfg          config.Client
	approver     *filter.Approver
	destinations map[string]config.Allow
}

// NewInbound returns a new Inbound.
func NewInbound(cfg config.Client, l logger.Logger) (in *Inbound, err error) {
	in = &Inbound{
		cfg: cfg,
		log: l,
	}

	var stricts, cidrs []string
	for _, allowed := range cfg.AllowList {
		if allowed.Type == "cidr" {
			cidrs = append(cidrs, allowed.Endpoint)
			continue
		}

		stricts = append(stricts, allowed.Endpoint)
	}

	in.approver, err = filter.NewAppover(stricts, cidrs)
	if err != nil {
		return in, err
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
		Source:       "remote_side",
		Remote:       in.cfg.Server.Address,
		Destination:  destination,
		IgnoreErrors: in.destinations[destination].IgnoreErrorsRegexp,
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

func (in *Inbound) getDestinations() (map[string]config.Allow, error) {
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

	// Check if server's destinations on client host are allowed.

	m := make(map[string]config.Allow)
	for _, wanted := range resp.(*control.InboundsResp).Inbounds {
		err = in.approver.Allowed(context.Background(), wanted)
		if err != nil {
			in.log.WithError(err).Warnf("Dropped destination %s", wanted)
			continue
		}

		// Here we are keeping allowed destinations for the next loop.
		m[wanted] = config.Allow{}
	}

	for _, allowed := range in.cfg.AllowList {
		if _, ok := m[allowed.Endpoint]; ok {
			// We overrides with allows that contains the IgnoreErrors patterns loaded from configuration.
			m[allowed.Endpoint] = allowed.Allow
		}
	}

	return m, errors.Wrap(err, "failed to get destinations")
}
