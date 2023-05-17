package filter

import (
	"context"
	"fmt"
	"net/url"

	"github.com/pkg/errors"
)

// ErrHostNotAllowed is returned when a host is not allowed.
var ErrHostNotAllowed = errors.New("host not allowed")

// An Approver is able to check if an host is allowed or not.
type Approver struct {
	resolver NameResolver
	stricts  []string
}

// NewAppover returns a new Approver where `stricts' is a slice od allowed hosts and cidrs a slice of allowed IP ranges.
func NewAppover(stricts []string, cidrs []string) (*Approver, error) {
	r, err := NewNameResolver(cidrs)
	if err != nil {
		return nil, err
	}

	return &Approver{
		resolver: *r,
		stricts:  stricts,
	}, nil
}

// Allowed checks if the host `host:port' is allowed or not.
// Use `errors.Is(err, ErrHostNotAllowed)' to check the error's nature.
func (f *Approver) Allowed(ctx context.Context, host string) error {
	for _, allowed := range f.stricts {
		if host == allowed {
			return nil
		}
	}

	// Resolve returns an error if there is no matching CIDRs for the given hostname.
	_, _, err := f.resolver.Resolve(ctx, (&url.URL{Host: host}).Hostname())
	if err != nil {
		return fmt.Errorf("%w: %w", ErrHostNotAllowed, err)
	}

	return nil
}
