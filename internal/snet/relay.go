package snet

import (
	"io"
	"net"
	"sync"
	"time"
)

// Relay copies between local and remote bidirectionally. Returns number of
// bytes copied from remote to local, from local to remote, and any error occurred.
// Borrowed from: https://github.com/shadowsocks/go-shadowsocks2
func Relay(local, remote net.Conn) error {
	var err, err1 error
	var wg sync.WaitGroup
	delay := time.Second

	// local = Dumper(local)
	// remote = Dumper(remote)

	wg.Add(1)
	go func() {
		defer wg.Done()

		_, err1 = io.Copy(remote, local)
		remote.SetDeadline(time.Now().Add(delay)) // wake up the other goroutine blocking on remote
	}()

	_, err = io.Copy(local, remote)
	local.SetDeadline(time.Now().Add(delay)) // wake up the other goroutine blocking on local

	wg.Wait()

	if err1 != nil {
		return err1
	}
	return err
}
