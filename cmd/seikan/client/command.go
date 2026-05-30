package client

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"regexp"

	"github.com/mdouchement/logger"
	"github.com/mdouchement/seikan/internal/client"
	"github.com/mdouchement/seikan/internal/config"
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

			level := slog.LevelInfo
			if cfg.Log.Level != "" {
				level, err = logger.ParseSlogLevel(cfg.Log.Level)
				if err != nil {
					return err
				}
				fmt.Println("Log level:", cfg.Log.Level)
			}

			l := slog.New(logger.NewSlogTextHandler(os.Stdout, &logger.SlogTextOption{
				Level:           level,
				DisableColors:   !cfg.Log.ForceColor,
				ForceColors:     cfg.Log.ForceColor,
				ForceFormatting: cfg.Log.ForceFormating,
				PrefixRE:        regexp.MustCompile(`^(\[.*?\])\s`),
				FullTimestamp:   true,
				TimestampFormat: "2006-01-02 15:04:05",
			}))

			client := client.New(cfg, logger.WrapSlog(l))
			err = client.Dial()
			if err != nil {
				return err
			}

			ctx, cancel := signal.NotifyContext(c.Context(), os.Interrupt)
			defer cancel()
			<-ctx.Done()
			return nil
		},
	}

	c.Flags().StringVarP(&filename, "config", "c", "", "Configuration file")
	return c
}
