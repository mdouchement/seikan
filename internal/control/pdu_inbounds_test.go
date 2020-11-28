package control_test

import (
	"bytes"
	"testing"

	"github.com/mdouchement/basex"
	"github.com/mdouchement/seikan/internal/control"
	"github.com/stretchr/testify/assert"
)

func TestInbouds(t *testing.T) {
	var pdu control.PDU = control.NewInbounds()
	pdu.RawHeader().SetSize(42)

	assert.Equal(t, 42, pdu.Size())
	assert.Equal(t, 0x01, pdu.Version())
	assert.Equal(t, control.InboundsID, pdu.ControlID())
	assert.NotEmpty(t, pdu.PID())
}

func TestInboudsSerialization(t *testing.T) {
	input := control.NewInbounds()
	input.Identifier = "id-1"

	p, err := control.Encode(input)
	assert.NoError(t, err)

	output, err := control.Decode(bytes.NewBuffer(p))
	assert.NoError(t, err)

	assert.Equal(t, input, output)
}

func TestInboudsResp(t *testing.T) {
	id := basex.GenerateID()
	var pdu control.PDU = control.NewInboundsResp(id) // interface compliance
	pdu.RawHeader().SetSize(42)

	assert.Equal(t, 42, pdu.Size())
	assert.Equal(t, 0x01, pdu.Version())
	assert.Equal(t, control.InboundsRespID, pdu.ControlID())
	assert.Equal(t, id, pdu.PID())
}

func TestInboudsRespSerialization(t *testing.T) {
	input := control.NewInboundsResp("unique-id")
	input.Inbounds = []string{"@1", "@2"}

	p, err := control.Encode(input)
	assert.NoError(t, err)

	output, err := control.Decode(bytes.NewBuffer(p))
	assert.NoError(t, err)

	assert.Equal(t, input, output)
}
