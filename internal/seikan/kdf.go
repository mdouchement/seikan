package seikan

import (
	"crypto/rand"
	"crypto/subtle"
	"hash"
	"io"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/hkdf"
)

// KDF global configuration.
const (
	KDFLength  = 48
	saltlength = 16
)

// KDFGenerate generates a safe derivation hash of the given s.
func KDFGenerate(s string) ([]byte, error) {
	payload := make([]byte, KDFLength)
	if _, err := io.ReadFull(rand.Reader, payload[:saltlength]); err != nil {
		return nil, err
	}

	nash := func() hash.Hash {
		h, err := blake2b.New256(payload[:saltlength])
		if err != nil {
			panic(err)
		}
		return h
	}

	kdf := hkdf.New(nash, []byte(s), nil, nil)
	_, err := io.ReadFull(kdf, payload[saltlength:])
	return payload, err
}

// KDFCompare returns true if the hased payload match the given s.
func KDFCompare(payload []byte, s string) bool {
	nash := func() hash.Hash {
		h, err := blake2b.New256(payload[:saltlength])
		if err != nil {
			panic(err)
		}
		return h
	}
	kdf := hkdf.New(nash, []byte(s), nil, nil)

	hash := make([]byte, KDFLength-saltlength)
	_, err := io.ReadFull(kdf, hash)
	if err != nil {
		panic(err)
	}

	return subtle.ConstantTimeCompare(payload[saltlength:], hash) == 1
}
