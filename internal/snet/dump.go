package snet

import (
	"encoding/hex"
	"net"
	"os"
)

type dumper struct {
	net.Conn
	hexdump bool
}

// Dumper dumps the given connection to stdout in raw or hexdump formats.
func Dumper(c net.Conn, hexdump ...bool) net.Conn {
	return &dumper{
		Conn:    c,
		hexdump: len(hexdump) > 0 && hexdump[0],
	}
}

func (d *dumper) Read(p []byte) (n int, err error) {
	n, err = d.Conn.Read(p)
	d.dump(p[:n])
	return n, err
}

func (d *dumper) Write(p []byte) (n int, err error) {
	d.dump(p)
	return d.Conn.Write(p)
}

func (d *dumper) Close() error {
	return d.Conn.Close()
}

func (d *dumper) dump(p []byte) {
	if d.hexdump && len(p) > 0 {
		hex.Dump(p)
		return
	}
	os.Stdout.Write(p)
}
