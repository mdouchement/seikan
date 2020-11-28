package control_test

import (
	"bytes"
	"testing"

	"github.com/mdouchement/basex"
	"github.com/mdouchement/seikan/internal/control"
	"github.com/stretchr/testify/assert"
)

func TestBindCS(t *testing.T) {
	var pdu control.PDU = control.NewBindCS()
	pdu.RawHeader().SetSize(42)

	assert.Equal(t, 42, pdu.Size())
	assert.Equal(t, 0x01, pdu.Version())
	assert.Equal(t, control.BindCSID, pdu.ControlID())
	assert.NotEmpty(t, pdu.PID())
}

func TestBindCSSerialization(t *testing.T) {
	input := control.NewBindCS()
	input.Identifier = "id-1"
	input.Address = "@"

	p, err := control.Encode(input)
	assert.NoError(t, err)

	output, err := control.Decode(bytes.NewBuffer(p))
	assert.NoError(t, err)

	assert.Equal(t, input, output)
}

func TestBindCSResp(t *testing.T) {
	id := basex.GenerateID()
	var pdu control.PDU = control.NewBindCSResp(id) // interface compliance
	pdu.RawHeader().SetSize(42)

	assert.Equal(t, 42, pdu.Size())
	assert.Equal(t, 0x01, pdu.Version())
	assert.Equal(t, control.BindCSRespID, pdu.ControlID())
	assert.Equal(t, id, pdu.PID())
}

func TestBindCSRespSerialization(t *testing.T) {
	input := control.NewBindCSResp("unique-id")

	p, err := control.Encode(input)
	assert.NoError(t, err)

	output, err := control.Decode(bytes.NewBuffer(p))
	assert.NoError(t, err)

	assert.Equal(t, input, output)
}
