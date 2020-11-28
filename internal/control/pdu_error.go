package control

import "fmt"

// Error is the payload in case of error.
type Error struct {
	*Header `cbor:"-"`
	Status  int    `cbor:"status"`
	Message string `cbor:"message"`
}

// NewError returns a new Error.
func NewError(id string) *Error {
	return &Error{
		Header: &Header{
			version: 0x01,
			cid:     ErrorID,
			pid:     id,
		},
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.Status, e.Message)
}
