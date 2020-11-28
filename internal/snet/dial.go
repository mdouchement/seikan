package snet

import (
	"errors"
	"net"
	"net/url"

	mnet "github.com/mendsley/gomnet"
)

// EnableKeepAlive enables keep alive if the connection allows it.
func EnableKeepAlive(c net.Conn) {
	if tcpc, ok := c.(*net.TCPConn); ok {
		tcpc.SetKeepAlive(true)
	}
}

// Listen announces on the local network address.
func Listen(netaddr string) (net.Listener, error) {
	url, err := url.Parse(netaddr)
	if err != nil {
		return nil, err
	}

	switch url.Scheme {
	case "tcp", "tcp4", "tcp6":
		return net.Listen(url.Scheme, url.Host)
	case "mnet":
		// Experitmental support of reliable-ordered transport over udp.
		return mnet.Listen("udp", url.Host)
	}

	return nil, errors.New("unsupported protocol scheme")
}

// Dial connects to the address on the named network.
func Dial(netaddr string) (net.Conn, error) {
	url, err := url.Parse(netaddr)
	if err != nil {
		return nil, err
	}

	switch url.Scheme {
	case "tcp", "tcp4", "tcp6":
		return net.Dial(url.Scheme, url.Host)
	case "mnet":
		// Experitmental support of reliable-ordered transport over udp.
		return mnet.Dial("udp", url.Host)
	}

	return nil, errors.New("unsupported protocol scheme")
}
