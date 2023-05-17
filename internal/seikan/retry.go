package seikan

import (
	"errors"
	"time"
)

// ErrNotRetayable is used when it's not used to perform a retry.
var ErrNotRetayable = errors.New("not retryable")

// Retry applies an exponential backoff if h returns an error.
func Retry(h func(error) error) {
	delay := 100 * time.Millisecond
	max := 20 * time.Second
	var err error

	for {
		err = h(err)
		if errors.Is(err, ErrNotRetayable) {
			return
		}

		//
		time.Sleep(delay)
		delay *= 2
		if delay > max {
			delay = max
		}
	}
}

// IsRetryNewError returns true if the err is different from prev.
func IsRetryNewError(prev, err error) bool {
	if err == nil {
		return false
	}

	if err != nil && prev == nil {
		return true
	}

	return prev.Error() != prev.Error()
}
