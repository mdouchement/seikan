package client

import (
	"fmt"
	"os"
	"os/signal"
	"regexp"

	"github.com/mdouchement/logger"
	"github.com/mdouchement/seikan/internal/client"
	"github.com/mdouchement/seikan/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Command starts the client.
func Command() *cobra.Command {
	var filename string

	c := &cobra.Command{
		Use:   "client",
		Short: "Start client",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) (err error) {
			var cfg config.Client
			err = config.Load(filename, &cfg)
			if err != nil {
				return err
			}

			l := logrus.New()
			if cfg.Log.Level != "" {
				level, err := logrus.ParseLevel(cfg.Log.Level)
				if err != nil {
					return err
				}
				fmt.Println("Log level:", cfg.Log.Level)
				l.SetLevel(level)
			}
			l.SetFormatter(&logger.LogrusTextFormatter{
				DisableColors:   !cfg.Log.ForceColor,
				ForceColors:     cfg.Log.ForceColor,
				ForceFormatting: cfg.Log.ForceFormating,
				PrefixRE:        regexp.MustCompile(`^(\[.*?\])\s`),
				FullTimestamp:   true,
				TimestampFormat: "2006-01-02 15:04:05",
			})

			client := client.New(cfg, logger.WrapLogrus(l))
			err = client.Dial()
			if err != nil {
				return err
			}

			signals := make(chan os.Signal, 1)
			signal.Notify(signals, os.Interrupt, os.Kill)
			<-signals
			return nil
		},
	}

	c.Flags().StringVarP(&filename, "config", "c", "", "Configuration file")
	return c
}
