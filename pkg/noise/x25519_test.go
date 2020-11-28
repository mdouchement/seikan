package noise_test

import (
	"crypto/rand"
	"testing"

	"github.com/mdouchement/seikan/pkg/base58"
	"github.com/mdouchement/seikan/pkg/noise"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/nacl/box"
)

func TestNewX25519String(t *testing.T) {
	i := noise.GenerateX25519Identity()
	assert.Equal(t, base58.Encode(i.PrivateKey()), i.PrivateKeyString())
	assert.Equal(t, base58.Encode(i.PublicKey()), i.PublicKeyString())
}

func TestParseX25519Recipient(t *testing.T) {
	i := noise.GenerateX25519Identity()
	recipient := noise.ParseX25519Recipient(i.PublicKeyString())
	assert.Equal(t, i.PublicKey(), *recipient)
}

func TestParseX25519key(t *testing.T) {
	i := noise.GenerateX25519Identity()

	key := make([]byte, curve25519.ScalarSize)
	noise.ParseX25519key(i.PrivateKeyString(), key)
	assert.Equal(t, []byte(i.PrivateKey()), key)
}

func TestParseX25519Identity(t *testing.T) {
	i := noise.GenerateX25519Identity()
	ni := noise.ParseX25519Identity(i.PrivateKeyString(), i.PublicKeyString())
	assert.Equal(t, i, ni)
}

func TestGenerateX25519Identity(t *testing.T) {
	alice := noise.GenerateX25519Identity()
	bob := noise.GenerateX25519Identity()

	checkX25519ECDH(t, alice, bob)
	checkX25519Encryption(t, alice, bob)
}

func TestNewX25519FromScalar(t *testing.T) {
	aliceScalar := make([]byte, curve25519.ScalarSize)
	_, err := rand.Read(aliceScalar)
	assert.NoError(t, err)
	alice := noise.NewX25519FromScalar(aliceScalar)

	bobScalar := make([]byte, curve25519.ScalarSize)
	_, err = rand.Read(bobScalar)
	assert.NoError(t, err)
	bob := noise.NewX25519FromScalar(bobScalar)

	checkX25519ECDH(t, alice, bob)
	checkX25519Encryption(t, alice, bob)
}

func checkX25519ECDH(t *testing.T, alice, bob *noise.X25519Identity) {
	assert.Len(t, alice.PrivateKey(), curve25519.ScalarSize)
	assert.Len(t, alice.PublicKey(), curve25519.ScalarSize)

	assert.Len(t, bob.PrivateKey(), curve25519.ScalarSize)
	assert.Len(t, bob.PublicKey(), curve25519.ScalarSize)

	//

	ab, err := curve25519.X25519(alice.PrivateKey(), bob.PublicKey())
	assert.NoError(t, err)
	ba, err := curve25519.X25519(bob.PrivateKey(), alice.PublicKey())
	assert.NoError(t, err)

	assert.Equal(t, ab, ba)
}

func checkX25519Encryption(t *testing.T, alice, bob *noise.X25519Identity) {
	assert.Len(t, alice.PrivateKey(), curve25519.ScalarSize)
	assert.Len(t, alice.PublicKey(), curve25519.ScalarSize)

	assert.Len(t, bob.PrivateKey(), curve25519.ScalarSize)
	assert.Len(t, bob.PublicKey(), curve25519.ScalarSize)

	//

	var privateAlice [32]byte
	copy(privateAlice[:], alice.PrivateKey())
	var publicAlice [32]byte
	copy(publicAlice[:], alice.PublicKey())

	var privateBob [32]byte
	copy(privateBob[:], bob.PrivateKey())
	var publicBob [32]byte
	copy(publicBob[:], bob.PublicKey())

	//

	message := []byte("message")
	var nonce [24]byte
	_, err := rand.Read(nonce[:])
	assert.NoError(t, err)

	seal := box.Seal(nil, message, &nonce, &publicBob, &privateAlice)
	msg, ok := box.Open(nil, seal, &nonce, &publicAlice, &privateBob)

	assert.True(t, ok)
	assert.Equal(t, message, msg)
}
