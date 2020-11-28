package main

import (
	"fmt"
	"log"

	"github.com/mdouchement/seikan/cmd/seikan/client"
	"github.com/mdouchement/seikan/cmd/seikan/identity"
	"github.com/mdouchement/seikan/cmd/seikan/server"
	"github.com/spf13/cobra"
)

var (
	version  = "dev"
	revision = "none"
	date     = "unknown"
)

func main() {
	c := &cobra.Command{
		Use:     "seikan",
		Short:   "TCP tunnels leveraging Noise Protocol",
		Version: fmt.Sprintf("%s - build %.7s @ %s", version, revision, date),
		Args:    cobra.NoArgs,
	}
	c.AddCommand(server.Command())
	c.AddCommand(client.Command())
	c.AddCommand(identity.Command())
	c.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Version for Seikan",
		Args:  cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(c.Version)
		},
	})

	if err := c.Execute(); err != nil {
		log.Fatal(err)
	}
}
