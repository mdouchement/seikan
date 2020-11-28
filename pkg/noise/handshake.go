package noise

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/flynn/noise"
	"golang.org/x/crypto/curve25519"
)

// An HandshakeOptions describes all the options needed for a Noise Protocol handshake.
type HandshakeOptions struct {
	Pattern HandshakePattern
	pattern noise.HandshakePattern

	Hash HashFunction
	hash noise.HashFunc

	Cipher   CipherFunction
	cipher   noise.CipherFunc
	overhead int

	//

	Sender    *X25519Identity
	Recipient *X25519Recipient
}

// Validate checks if the options are valid.
func (o *HandshakeOptions) Validate() error {
	switch o.Pattern {
	case PatternIK:
		o.pattern = noise.HandshakeIK
	default:
		return errors.New("unsupported pattern")
	}

	//

	switch o.Hash {
	case HashBlake2b:
		o.hash = noise.HashBLAKE2b
	case HashBlake2s:
		o.hash = noise.HashBLAKE2s
	default:
		return errors.New("unsupported hash function")
	}

	//

	switch o.Cipher {
	case CipherChaCha20Poly1305:
		o.overhead = 16 // Taken from a new instance of ChaCha20-Poly1305
		o.cipher = new(chacha20poly1305fn)
	case CipherAES256GCM:
		o.overhead = 16 // Taken from a new instance of AES256-GCM
		o.cipher = noise.CipherAESGCM
	default:
		return errors.New("unsupported cipher function")
	}

	//

	if o.Sender == nil {
		return errors.New("sender not provided")
	}

	if len(o.Sender.PrivateKey()) != curve25519.ScalarSize {
		return errors.New("invalid sender private key")
	}

	if len(o.Sender.PublicKey()) != curve25519.ScalarSize {
		return errors.New("invalid sender public key")
	}

	//

	if o.Recipient == nil {
		return errors.New("recipient not provided")
	}

	if len(*o.Recipient) != curve25519.ScalarSize {
		fmt.Println(o.Recipient)
		return errors.New("invalid recipient")
	}

	return nil
}

/////////////////////
//                 //
// Handshake       //
//                 //
/////////////////////

type handshake struct {
	initiator bool
	pattern   noise.HandshakePattern
	state     *noise.HandshakeState
	stream    io.ReadWriter
	overhead  int
	buf       []byte
}

// Handshake performs the handshake for X25519Identity sender and recipient using given stream.
//
// The cipher suite used is:
// Curve25519 ECDH and provided Cipher and Hash.
//
// One of the Noise participants should be the initiator.
//
// Documentation:
// https://noiseprotocol.org/noise.html
// https://latacora.micro.blog/factoring-the-noise/
func Handshake(stream io.ReadWriter, options HandshakeOptions, initiator bool) (Cipher, error) {
	if err := options.Validate(); err != nil {
		return nil, err
	}

	dhKey := noise.DHKey{
		Private: options.Sender.PrivateKey(),
		Public:  options.Sender.PublicKey(),
	}

	prePeerStatic := (initiator && options.pattern.InitiatorPreMessages == nil) || (!initiator && options.pattern.ResponderPreMessages == nil)

	config := noise.Config{
		CipherSuite:   noise.NewCipherSuite(noise.DH25519, options.cipher, options.hash),
		Pattern:       options.pattern,
		Initiator:     initiator,
		Prologue:      []byte("seikan/1.0"),
		StaticKeypair: dhKey,
	}
	if prePeerStatic {
		config.PeerStatic = *options.Recipient
	}

	state, err := noise.NewHandshakeState(config)
	if err != nil {
		return nil, err
	}

	h := &handshake{
		initiator: initiator,
		pattern:   options.pattern,
		state:     state,
		stream:    stream,
		overhead:  options.overhead,
		buf:       make([]byte, SizePrefixLength+noise.MaxMsgLen),
	}

	var cipher Cipher
	if h.initiator {
		cipher, err = h.begin(stream)
	} else {
		cipher, err = h.wait(stream)
	}

	if err != nil {
		return nil, err
	}

	if !prePeerStatic {
		err := errors.New("invalid peer")
		peer := h.state.PeerStatic()

		if len(peer) != len(*options.Recipient) {
			return nil, err
		}

		for i, b := range *options.Recipient {
			if peer[i] != b {
				return nil, err
			}
		}
	}

	return cipher, nil
}

// Finished indicate whether handshake is completed.
func (h *handshake) Finished() bool {
	return h.state.MessageIndex() == len(h.pattern.Messages)
}

func (h *handshake) begin(stream io.ReadWriter) (Cipher, error) {
	var err error
	cipher := &cipher{
		overhead: h.overhead,
	}

	for {
		cipher.tr, cipher.tx, err = h.write()
		if err != nil {
			return nil, err
		}
		if h.Finished() {
			return cipher, nil
		}

		cipher.tx, cipher.tr, err = h.read()
		if err != nil {
			return nil, err
		}
		if h.Finished() {
			return cipher, nil
		}
	}
}

func (h *handshake) wait(stream io.ReadWriter) (Cipher, error) {
	var err error
	cipher := &cipher{
		overhead: h.overhead,
	}

	for {
		cipher.tx, cipher.tr, err = h.read()
		if err != nil {
			return nil, err
		}
		if h.Finished() {
			return cipher, nil
		}

		cipher.tr, cipher.tx, err = h.write()
		if err != nil {
			return nil, err
		}
		if h.Finished() {
			return cipher, nil
		}
	}
}

func (h *handshake) read() (*noise.CipherState, *noise.CipherState, error) {
	buf := h.buf[:]

	// Read message length
	_, err := h.stream.Read(buf[:SizePrefixLength])
	if err != nil {
		return nil, nil, err
	}
	n := binary.BigEndian.Uint16(buf[:SizePrefixLength])

	// Read message
	_, err = h.stream.Read(buf[:n])
	if err != nil {
		return nil, nil, err
	}

	_, cs0, cs1, err := h.state.ReadMessage(nil, buf[:n])
	return cs0, cs1, err
}

func (h *handshake) write() (*noise.CipherState, *noise.CipherState, error) {
	buf := h.buf[:SizePrefixLength]

	// Generate message
	message, cs0, cs1, err := h.state.WriteMessage(buf[SizePrefixLength:], nil)
	if err != nil {
		return cs0, cs1, err
	}
	n := len(message)

	// Prepend message length and send message
	binary.BigEndian.PutUint16(buf[:SizePrefixLength], uint16(n))
	_, err = h.stream.Write(buf[:SizePrefixLength+n])
	return cs0, cs1, err
}
