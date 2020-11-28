package noise_test

import (
	"io"
	"testing"

	"github.com/mdouchement/seikan/pkg/noise"
	"github.com/mdouchement/seikan/pkg/stream"
	"github.com/stretchr/testify/assert"
)

func TestHandshakeOptions_Validate(t *testing.T) {
	options := noise.HandshakeOptions{}
	assert.EqualError(t, options.Validate(), "unsupported pattern")

	options.Pattern = noise.PatternIK
	assert.EqualError(t, options.Validate(), "unsupported hash function")

	options.Hash = noise.HashBlake2b
	assert.EqualError(t, options.Validate(), "unsupported cipher function")

	options.Cipher = noise.CipherChaCha20Poly1305
	assert.EqualError(t, options.Validate(), "sender not provided")

	options.Sender = noise.GenerateX25519Identity()
	assert.EqualError(t, options.Validate(), "recipient not provided")

	public := noise.GenerateX25519Identity().PublicKey()
	options.Recipient = &public
	assert.NoError(t, options.Validate())

	options.Hash = noise.HashBlake2s
	options.Cipher = noise.CipherAES256GCM
	assert.NoError(t, options.Validate())
}

func TestHandshake(t *testing.T) {
	alice := noise.GenerateX25519Identity()
	bob := noise.GenerateX25519Identity()

	conn := stream.NewBidirectional()

	bob2alice := "trololo!"
	alice2bob := "popo!"

	go func() {
		recipient := bob.PublicKey()
		options := noise.HandshakeOptions{
			Pattern:   noise.PatternIK,
			Hash:      noise.HashBlake2b,
			Cipher:    noise.CipherChaCha20Poly1305,
			Sender:    alice,
			Recipient: &recipient,
		}

		c, err := noise.Handshake(conn.C1, options, false)
		assert.NoError(t, err, "alice")

		message, err := receive(conn.C1, c)
		assert.NoError(t, err, "alice")
		assert.Equal(t, bob2alice, message)

		err = send(conn.C1, c, alice2bob)
		assert.NoError(t, err, "alice")
	}()

	//

	recipient := alice.PublicKey()
	options := noise.HandshakeOptions{
		Pattern:   noise.PatternIK,
		Hash:      noise.HashBlake2b,
		Cipher:    noise.CipherChaCha20Poly1305,
		Sender:    bob,
		Recipient: &recipient,
	}

	c, err := noise.Handshake(conn.C2, options, true)
	assert.NoError(t, err, "bob")

	err = send(conn.C2, c, bob2alice)
	assert.NoError(t, err, "bob")

	message, err := receive(conn.C2, c)
	assert.NoError(t, err, "bob")
	assert.Equal(t, alice2bob, message)
}

func send(c io.ReadWriter, cipher noise.Cipher, message string) error {
	ciphertext, err := cipher.Encrypt(nil, nil, []byte(message))
	if err != nil {
		return err
	}

	_, err = c.Write(ciphertext)
	return err
}

func receive(c io.ReadWriter, cipher noise.Cipher) (string, error) {
	ciphertext := make([]byte, 4096)
	n, err := c.Read(ciphertext)
	if err != nil {
		return "", err
	}

	message, err := cipher.Decrypt(nil, nil, ciphertext[:n])
	return string(message), err
}
