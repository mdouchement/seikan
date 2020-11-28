package noise

import (
	cipherpkg "crypto/cipher"
	"encoding/binary"

	"github.com/flynn/noise"
	chacha20poly1305pkg "golang.org/x/crypto/chacha20poly1305"
)

// Implements noise.CipherFunc
type chacha20poly1305fn struct{}

func (chacha20poly1305fn) CipherName() string {
	return "ChaChaPoly"
}

func (chacha20poly1305fn) Cipher(k [32]byte) noise.Cipher {
	c, err := chacha20poly1305pkg.New(k[:])
	if err != nil {
		panic(err)
	}

	return chacha20poly1305{
		AEAD: c,
	}
}

// Implements noise.Cipher.
// It reuses nonce slice for better performance, bytes are rewritten each time by PutUint64.
type chacha20poly1305 struct {
	cipherpkg.AEAD
	nonce [12]byte
}

func (c chacha20poly1305) Encrypt(out []byte, n uint64, ad, plaintext []byte) []byte {
	binary.LittleEndian.PutUint64(c.nonce[4:], n)
	return c.Seal(out, c.nonce[:], plaintext, ad)
}

func (c chacha20poly1305) Decrypt(out []byte, n uint64, ad, ciphertext []byte) ([]byte, error) {
	binary.LittleEndian.PutUint64(c.nonce[4:], n)
	return c.Open(out, c.nonce[:], ciphertext, ad)
}
