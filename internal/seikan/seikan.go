package seikan

import (
	"strings"
)

// CraftKey builds a dot separated key of the given args.
func CraftKey(args ...string) string {
	return strings.Join(args, ".")
}
