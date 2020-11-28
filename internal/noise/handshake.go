package noise

import (
	"net"

	"github.com/mdouchement/seikan/pkg/noise"
	"github.com/pkg/errors"
)

// Handshake performs the handshake for Identity sender and recipient using the given net.Conn.
//
// The cipher suite used is:
// Curve25519 ECDH, ChaCha20-Poly1305 AEAD, BLAKE2b hash.
//
// The handshake uses the IK pattern:
// I = Static key for initiator Immediately transmitted to responder, despite reduced or absent identity hiding
// K = Static key for responder Known to initiator
//
// One of the Noise participants should be the initiator.
func Handshake(c net.Conn, sender Identity, recipient string, initiator bool) (net.Conn, error) {
	options := noise.HandshakeOptions{
		Pattern:   noise.PatternIK,
		Hash:      noise.HashBlake2b,
		Cipher:    noise.CipherChaCha20Poly1305,
		Sender:    noise.ParseX25519Identity(sender.private(), sender.public()),
		Recipient: noise.ParseX25519Recipient(Identity{Public: recipient}.public()),
	}

	cipher, err := noise.Handshake(c, options, initiator)
	if err != nil {
		return nil, errors.Wrap(err, "handshake")
	}

	return noise.NewChunkedConn(c, cipher), nil
}
