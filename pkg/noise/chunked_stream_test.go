package noise_test

import (
	"bytes"
	"crypto/rand"
	mrand "math/rand"
	"testing"
	"time"

	"github.com/mdouchement/seikan/pkg/noise"
	"github.com/mdouchement/seikan/pkg/stream"
	"github.com/stretchr/testify/assert"
)

func TestChunkedStream(t *testing.T) {
	rng := mrand.New(mrand.NewSource(time.Now().UnixNano()))

	//

	alice := noise.GenerateX25519Identity()
	bob := noise.GenerateX25519Identity()

	conn := stream.NewBidirectional()

	expected := make([]byte, 0xFFF0+rng.Intn(0xFFFF))
	_, err := rand.Read(expected)
	assert.NoError(t, err)

	p := make([]byte, 1024)
	actual := bytes.NewBuffer(nil)

	next := make(chan bool)

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

		s1 := noise.NewChunkedStream(conn.C1, c)

		_, err = s1.Write(expected)
		assert.NoError(t, err, "alice")

		<-next

		for actual.Len() < len(expected) {
			n, err := s1.Read(p)
			assert.NoError(t, err, "alice")
			_, err = actual.Write(p[:n])
			assert.NoError(t, err, "alice")
		}
		assert.Equal(t, expected, actual.Bytes(), "alice")

		next <- true
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

	s2 := noise.NewChunkedStream(conn.C2, c)

	for actual.Len() < len(expected) {
		n, err := s2.Read(p)
		assert.NoError(t, err, "bob")
		_, err = actual.Write(p[:n])
		assert.NoError(t, err, "bob")
	}
	assert.Equal(t, expected, actual.Bytes(), "bob")

	actual.Reset() // Important!

	next <- true

	_, err = s2.Write(expected)
	assert.NoError(t, err, "bob")

	<-next
}
