package noise

import "github.com/flynn/noise"

type (
	// Cipher provides symmetric encryption and decryption after a successful handshake.
	Cipher interface {
		Overhead() int
		EncryptRekey()
		Encrypt(out, ad, plaintext []byte) ([]byte, error)
		DecryptRekey()
		Decrypt(out, ad, ciphertext []byte) ([]byte, error)
	}

	cipher struct {
		overhead int
		tx       *noise.CipherState
		tr       *noise.CipherState
	}
)

func (c *cipher) Overhead() int {
	return c.overhead
}

func (c *cipher) EncryptRekey() {
	c.tx.Rekey()
}

func (c *cipher) Encrypt(out, ad, plaintext []byte) ([]byte, error) {
	return c.tx.Encrypt(out, ad, plaintext)
}

func (c *cipher) DecryptRekey() {
	c.tr.Rekey()
}

func (c *cipher) Decrypt(out, ad, ciphertext []byte) ([]byte, error) {
	return c.tr.Decrypt(out, ad, ciphertext)
}
