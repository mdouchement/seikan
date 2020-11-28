package seikan_test

import (
	"encoding/hex"
	"testing"

	"github.com/mdouchement/basex"
	"github.com/mdouchement/seikan/internal/seikan"
	"github.com/stretchr/testify/assert"
)

func TestKDF(t *testing.T) {
	n := 20
	h := make(map[string]bool)
	id := basex.GenerateID()

	for i := 0; i < n; i++ {
		payload, err := seikan.KDFGenerate(id)
		assert.NoError(t, err)
		h[hex.EncodeToString(payload)] = true
	}
	assert.Len(t, h, n, "uniqueness")

	for v := range h {
		payload, err := hex.DecodeString(v)
		assert.NoError(t, err)
		assert.True(t, seikan.KDFCompare(payload, id), "end2end")
	}
}
