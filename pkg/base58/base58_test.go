package base58_test

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/mdouchement/seikan/pkg/base58"
	"github.com/stretchr/testify/assert"
)

func TestBase58(t *testing.T) {
	for i := 0; i < 20; i++ {
		b, err := rand.Int(rand.Reader, big.NewInt(100))
		assert.NoError(t, err)

		p := make([]byte, b.Int64())
		_, err = rand.Read(p)
		assert.NoError(t, err)

		s := base58.Encode(p)
		pp := base58.Decode(s)
		assert.Equal(t, p, pp)
	}
}
