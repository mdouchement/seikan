package control_test

import (
	"bytes"
	"testing"

	"github.com/mdouchement/basex"
	"github.com/mdouchement/seikan/internal/control"
	"github.com/stretchr/testify/assert"
)

func TestBindSC(t *testing.T) {
	var pdu control.PDU = control.NewBindSC()
	pdu.RawHeader().SetSize(42)

	assert.Equal(t, 42, pdu.Size())
	assert.Equal(t, 0x01, pdu.Version())
	assert.Equal(t, control.BindSCID, pdu.ControlID())
	assert.NotEmpty(t, pdu.PID())
}

func TestBindSCSerialization(t *testing.T) {
	input := control.NewBindSC()
	input.Identifier = "id-1"
	input.Address = "@"

	p, err := control.Encode(input)
	assert.NoError(t, err)

	output, err := control.Decode(bytes.NewBuffer(p))
	assert.NoError(t, err)

	assert.Equal(t, input, output)
}

func TestBindSCResp(t *testing.T) {
	id := basex.GenerateID()
	var pdu control.PDU = control.NewBindSCResp(id) // interface compliance
	pdu.RawHeader().SetSize(42)

	assert.Equal(t, 42, pdu.Size())
	assert.Equal(t, 0x01, pdu.Version())
	assert.Equal(t, control.BindSCRespID, pdu.ControlID())
	assert.Equal(t, id, pdu.PID())
}

func TestBindSCRespSerialization(t *testing.T) {
	input := control.NewBindSCResp("unique-id")

	p, err := control.Encode(input)
	assert.NoError(t, err)

	output, err := control.Decode(bytes.NewBuffer(p))
	assert.NoError(t, err)

	assert.Equal(t, input, output)
}
