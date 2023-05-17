package server

import (
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/mdouchement/basex"
	"github.com/mdouchement/logger"
	"github.com/mdouchement/seikan/internal/config"
	"github.com/mdouchement/seikan/internal/control"
	"github.com/mdouchement/seikan/internal/filter"
	"github.com/mdouchement/seikan/internal/noise"
	"github.com/mdouchement/seikan/internal/seikan"
	"github.com/mdouchement/seikan/internal/smux"
	"github.com/mdouchement/seikan/internal/snet"
	"github.com/pkg/errors"
)

type (
	// A Server listens on a port for running Seikan's tunnels.
	Server interface {
		Listen() error
	}

	server struct {
		cfg      config.Server
		log      logger.Logger
		approver *filter.Approver
		outbound *Outbound
	}

	stream func(c net.Conn) error
)

// New returns a new server.
func New(cfg config.Server, l logger.Logger) (Server, error) {
	s := &server{
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

	var err error
	s.approver, err = filter.NewAppover(stricts, cidrs)
	if err != nil {
		return s, err
	}

	s.outbound, err = NewOutbound(cfg, l)
	return s, err
}

// Listen listens for incoming tunnels.
func (s *server) Listen() error {
	l, err := snet.Listen(s.cfg.Address)
	if err != nil {
		return errors.Wrapf(err, "failed to listen on %s", s.cfg.Address)
	}

	s.log.Infof("Listening on %s", s.cfg.Address)
	for {
		c, err := l.Accept()
		if err != nil {
			s.log.WithError(err).Warn("failed to accept connection")
			continue
		}

		go func() {
			defer c.Close()
			snet.EnableKeepAlive(c)

			log := s.log.WithPrefixf("[%s]", basex.GenerateID())
			log.Info("Handshake")

			defer drain(log, c)

			//
			// Identifier
			//

			log.Debug("Reading derived identifier")
			derived := make([]byte, seikan.KDFLength)
			if _, err = io.ReadFull(c, derived); err != nil {
				log.Errorf("failed to read derived identifier: %s", err.Error())
				return
			}

			recipient, sessid, err := s.recipient(derived)
			if err != nil {
				log.Error(err.Error())
				return
			}

			//
			// Handshake
			//

			log.Debug("Performing Noise handshake")
			c, err = noise.Handshake(c, identity(s.cfg), recipient, true)
			if err != nil {
				if errors.Cause(err) != io.EOF {
					log = log.WithError(err)
				}
				log.Error("failed to perform handshake")
				return
			}

			//
			// Compression
			//

			c, err = snet.Compress(c)
			if err != nil {
				log.Error(err.Error())
				return
			}
			defer c.Close() // Compression only

			//
			// Controls
			//

			var await bool
			var stream stream
			for {
				pdu, err := control.Decode(c)
				if errors.Is(err, io.EOF) {
					log.Error("connection closed during control")
				}

				if err != nil {
					log.WithError(err).Error("failed to receive control")

					pid := "unkown"
					if pdu != nil {
						pid = pdu.PID()
					}

					resp := control.NewError(pid)
					resp.Status = http.StatusInternalServerError
					resp.Message = "failed to receive control"
					control.EncodeTo(c, resp)

					return
				}

				pdu, stream, await = s.control(log, sessid, pdu)
				if err = control.EncodeTo(c, pdu); err != nil {
					log.WithError(err).Error("failed to send control")

					resp := control.NewError(pdu.PID())
					resp.Status = http.StatusInternalServerError
					resp.Message = "failed to send control"
					control.EncodeTo(c, resp)

					return
				}

				if stream != nil {
					break
				}

				if !await {
					log.Info("closing connection")
					return
				}
			}

			//
			// Streaming
			//

			if err = stream(c); err != nil {
				log.WithError(err).Error("stream closed")
			}
		}()
	}
}

func (s *server) control(log logger.Logger, sessid string, pdu control.PDU) (control.PDU, stream, bool) {
	log.Infof("Performing %s control", pdu.ControlID())

	switch p := pdu.(type) {
	case *control.Inbounds:
		if p.Identifier == "" {
			resp := control.NewError(pdu.PID())
			resp.Status = http.StatusUnprocessableEntity
			resp.Message = "missing indentifier"

			return resp, nil, true
		}

		if p.Identifier != sessid {
			log.Warnf("Forbidden %s", p.Identifier)

			resp := control.NewError(pdu.PID())
			resp.Status = http.StatusForbidden
			resp.Message = "invalid identifier"

			return resp, nil, true
		}

		//

		var addresses []string
		for _, outbound := range s.cfg.Outbounds {
			if outbound.Identifier == p.Identifier {
				addresses = append(addresses, outbound.Destination)
			}
		}

		resp := control.NewInboundsResp(pdu.PID())
		resp.Inbounds = addresses

		return resp, nil, false
		//
		//
	case *control.BindCS:
		if p.Identifier == "" || p.Address == "" {
			resp := control.NewError(pdu.PID())
			resp.Status = http.StatusUnprocessableEntity
			resp.Message = "missing indentifier or address"

			return resp, nil, false
		}

		if p.Identifier != sessid {
			log.Warnf("Forbidden %s", p.Identifier)

			resp := control.NewError(pdu.PID())
			resp.Status = http.StatusForbidden
			resp.Message = "invalid identifier"

			return resp, nil, false
		}

		err := s.approver.Allowed(context.Background(), p.Address)
		if len(s.cfg.AllowList) > 0 && err != nil {
			log.WithError(err).Warnf("Rejected %s", p.Address)

			resp := control.NewError(pdu.PID())
			resp.Status = http.StatusUnprocessableEntity
			resp.Message = "rejected indentifier or address"

			return resp, nil, false
		}

		//

		tun := smux.Tunnel{
			Source:      p.Identifier,
			Remote:      s.cfg.Address,
			Destination: p.Address,
		}

		stream := func(c net.Conn) error {
			smux, err := smux.NewServer(log.WithPrefix("[ingoing ]"), tun, c)
			if err != nil {
				return errors.Wrapf(err, "failed to establish connection for %s.%s", p.Identifier, p.Address)
			}
			defer smux.Close()

			return smux.Listen()
		}

		//

		resp := control.NewBindCSResp(pdu.PID())
		return resp, stream, false
		//
		//
	case *control.BindSC:
		if p.Identifier == "" || p.Address == "" {
			resp := control.NewError(pdu.PID())
			resp.Status = http.StatusUnprocessableEntity
			resp.Message = "missing indentifier or address"

			return resp, nil, true
		}

		if p.Identifier != sessid {
			log.Warnf("Forbidden %s", p.Identifier)

			resp := control.NewError(pdu.PID())
			resp.Status = http.StatusForbidden
			resp.Message = "invalid identifier"

			return resp, nil, true
		}

		//

		stream := func(c net.Conn) error {
			return s.outbound.Establish(
				log.WithPrefix("[outgoing]"),
				config.Outbound{Identifier: p.Identifier, Destination: p.Address},
				c,
			)
		}

		//

		resp := control.NewBindSCResp(pdu.PID())
		return resp, stream, false
		//
		//
	default:
		resp := control.NewError(pdu.PID())
		resp.Status = http.StatusBadRequest
		resp.Message = "unsupported PDU"

		return resp, nil, true
	}
}

func (s *server) recipient(derived []byte) (string, string, error) {
	for identifier, receipient := range s.cfg.Clients {
		if seikan.KDFCompare(derived, identifier) {
			return receipient, identifier, nil
		}
	}

	return "", "", errors.New("unknown receipient")
}

func identity(c config.Server) noise.Identity {
	return noise.Identity{
		Secret: c.Secret,
		Public: c.Public,
	}
}

// Drain c to avoid leaking server behavioral features
// see https://www.ndss-symposium.org/ndss-paper/detecting-probe-resistant-proxies/
func drain(log logger.Logger, c net.Conn) {
	_, err := io.Copy(ioutil.Discard, c)
	if err != nil {
		log.Warnf("Draining error: %s", err)
	}
}
