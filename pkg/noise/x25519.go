package noise

import (
	"crypto/rand"

	"github.com/mdouchement/seikan/pkg/base58"
	"golang.org/x/crypto/curve25519"
)

type (
	// A X25519Identity is a X25519 assymmetric encryption key.
	X25519Identity struct {
		private X25519PrivateKey
		public  X25519PublicKey
	}

	// A X25519PrivateKey is the private part of a X25519 assymmetric encryption key.
	X25519PrivateKey []byte
	// A X25519PublicKey is the public part of a X25519 assymmetric encryption key.
	X25519PublicKey []byte
	// X25519Recipient is the standard X25519 public key, based on a Curve25519 point.
	X25519Recipient = X25519PublicKey
)

// ParseX25519Recipient returns a new X25519Recipient from a base58 public key.
func ParseX25519Recipient(s string) *X25519Recipient {
	var recipient X25519Recipient = make([]byte, curve25519.ScalarSize)
	ParseX25519key(s, recipient)
	return &recipient
}

// ParseX25519key parse the given s into key.
func ParseX25519key(s string, key []byte) {
	p := base58.Decode(s)
	copy(key, p[:curve25519.ScalarSize])
}

// ParseX25519Identity returns a new X25519Recipient from a base58 private and public keys.
func ParseX25519Identity(private, public string) *X25519Identity {
	identity := &X25519Identity{
		private: make([]byte, curve25519.ScalarSize),
		public:  make([]byte, curve25519.ScalarSize),
	}
	ParseX25519key(private, identity.private)
	ParseX25519key(public, identity.public)
	return identity
}

// GenerateX25519Identity returns a new X25519Identity.
func GenerateX25519Identity() *X25519Identity {
	scalar := make([]byte, curve25519.ScalarSize)
	if _, err := rand.Read(scalar); err != nil {
		panic(err)
	}

	return NewX25519FromScalar(scalar)
}

// NewX25519FromScalar returns a new X25519Identity based on the given 32 byte-length scalar.
func NewX25519FromScalar(scalar []byte) *X25519Identity {
	if len(scalar) != curve25519.ScalarSize {
		panic("bad scalar length")
	}

	identity := &X25519Identity{
		private: make([]byte, curve25519.ScalarSize),
	}
	copy(identity.private, scalar)
	identity.public, _ = curve25519.X25519(identity.private, curve25519.Basepoint)
	return identity
}

// PrivateKey returns the private key of the X25519 identity.
func (k *X25519Identity) PrivateKey() X25519PrivateKey {
	return k.private
}

// PrivateKeyString returns the encoded private key of the X25519 identity.
func (k *X25519Identity) PrivateKeyString() string {
	return base58.Encode(k.private[:])
}

// PublicKey returns the public key of the X25519 identity.
func (k *X25519Identity) PublicKey() X25519PublicKey {
	return k.public
}

// PublicKeyString returns the encoded public key of the X25519 identity.
func (k *X25519Identity) PublicKeyString() string {
	return base58.Encode(k.public[:])
}
