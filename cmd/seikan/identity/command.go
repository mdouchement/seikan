package identity

import (
	"fmt"

	"github.com/mdouchement/seikan/internal/noise"
	"github.com/spf13/cobra"
)

// Command generates a new identity.
func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "identity",
		Short: "Generate a new identity",
		Args:  cobra.NoArgs,
		Run: func(c *cobra.Command, _ []string) {
			i := noise.GenerateIdentity()
			fmt.Println("secret:", i.Secret)
			fmt.Println("public:", i.Public)
		},
	}
	return c
}
