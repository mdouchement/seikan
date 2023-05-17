package filter

import (
	"context"
	"net"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/pkg/errors"
)

// CacheTTL is the duration before a domain name resolution is evict form the cache.
const CacheTTL = 12 * time.Hour

// ErrHostRejected is returned when the host has been flagged as unwanted.
var ErrHostRejected = errors.New("rejected host")

// A NameResolver is used to filter IPs using name resolution.
type NameResolver struct {
	allows []*net.IPNet
	cache  *ristretto.Cache
}

// NewNameResolver return a new NameResolver.
func NewNameResolver(allows []string) (*NameResolver, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 50_000,
		MaxCost:     5000,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}

	resolver := &NameResolver{
		allows: make([]*net.IPNet, 0, len(allows)),
		cache:  cache,
	}

	for _, allow := range allows {
		_, block, err := net.ParseCIDR(allow)
		if err != nil {
			return nil, err
		}

		resolver.allows = append(resolver.allows, block)
	}

	return resolver, nil
}

// Resolve returns the ip for the given domain name.
func (r *NameResolver) Resolve(ctx context.Context, name string) (context.Context, net.IP, error) {
	if ip, ok := r.cache.Get(name); ok {
		return ctx, ip.(net.IP), nil
	}

	addr, err := net.ResolveIPAddr("ip", name)
	if err != nil {
		return ctx, nil, errors.Wrapf(err, "[resolve] %s", name)
	}

	for _, block := range r.allows {
		if block.Contains(addr.IP) {
			r.cache.SetWithTTL(name, addr.IP, 1, CacheTTL)
			return ctx, addr.IP, nil
		}
	}

	return ctx, nil, errors.Wrapf(ErrHostRejected, "[domain/ip] %s/%s", name, addr.IP)
}
