package noise

import (
	"strings"

	"github.com/mdouchement/seikan/pkg/noise"
)

// An Identity is a keypair for Diffie-Hellman calculation.
type Identity struct {
	Secret string
	Public string
}

// GenerateIdentity returns a new Identity.
func GenerateIdentity() Identity {
	i := noise.GenerateX25519Identity()
	return Identity{
		Secret: "sk-" + i.PrivateKeyString(),
		Public: "pk-" + i.PublicKeyString(),
	}
}

func (i Identity) private() string {
	return strings.ReplaceAll(i.Secret, "sk-", "")
}

func (i Identity) public() string {
	return strings.ReplaceAll(i.Public, "pk-", "")
}
