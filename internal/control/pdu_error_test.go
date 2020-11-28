package control_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/mdouchement/basex"
	"github.com/mdouchement/seikan/internal/control"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	id := basex.GenerateID()
	var pdu control.PDU = control.NewError(id) // interface compliance
	pdu.RawHeader().SetSize(42)

	assert.Equal(t, 42, pdu.Size())
	assert.Equal(t, 0x01, pdu.Version())
	assert.Equal(t, control.ErrorID, pdu.ControlID())
	assert.Equal(t, id, pdu.PID())
}

func TestErrorSerialization(t *testing.T) {
	input := control.NewError("unique-id")
	input.Status = http.StatusInternalServerError
	input.Message = "panic"

	p, err := control.Encode(input)
	assert.NoError(t, err)

	output, err := control.Decode(bytes.NewBuffer(p))
	assert.NoError(t, err)

	assert.Equal(t, input, output)
}
