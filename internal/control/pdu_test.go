package control_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/mdouchement/seikan/internal/control"
	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	pdu := control.NewError("unique-id")
	pdu.Status = http.StatusBadRequest
	pdu.Message = "msg"

	p, err := control.Encode(pdu)
	assert.NoError(t, err)

	expected := []byte{0, 37, 1, 1, 117, 110, 105, 113, 117, 101, 45, 105, 100, 0, 162, 102, 115, 116, 97, 116, 117, 115, 25, 1, 144, 103, 109, 101, 115, 115, 97, 103, 101, 99, 109, 115, 103}
	assert.Equal(t, expected, p)
}

func TestEncodeTo(t *testing.T) {
	pdu := control.NewError("unique-id")
	pdu.Status = http.StatusBadRequest
	pdu.Message = "msg"

	var buf bytes.Buffer

	err := control.EncodeTo(&buf, pdu)
	assert.NoError(t, err)

	expected := []byte{0, 37, 1, 1, 117, 110, 105, 113, 117, 101, 45, 105, 100, 0, 162, 102, 115, 116, 97, 116, 117, 115, 25, 1, 144, 103, 109, 101, 115, 115, 97, 103, 101, 99, 109, 115, 103}
	assert.Equal(t, expected, buf.Bytes())
}

func TestDecode(t *testing.T) {
	r := bytes.NewBuffer([]byte{0, 37, 1, 1, 117, 110, 105, 113, 117, 101, 45, 105, 100, 0, 162, 102, 115, 116, 97, 116, 117, 115, 25, 1, 144, 103, 109, 101, 115, 115, 97, 103, 101, 99, 109, 115, 103})
	pdu, err := control.Decode(r)
	assert.NoError(t, err)

	expected := control.NewError("unique-id")
	expected.RawHeader().SetSize(37)
	expected.Status = http.StatusBadRequest
	expected.Message = "msg"

	assert.Equal(t, expected, pdu)
}
