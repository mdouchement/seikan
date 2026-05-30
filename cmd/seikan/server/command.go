package server

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"

	"github.com/mdouchement/logger"
	"github.com/mdouchement/seikan/internal/config"
	"github.com/mdouchement/seikan/internal/server"
	"github.com/spf13/cobra"
)

// Command launches the server.
func Command() *cobra.Command {
	var filename string

	c := &cobra.Command{
		Use:   "server",
		Short: "Start server",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) (err error) {
			var cfg config.Server
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

			s, err := server.New(cfg, logger.WrapSlog(l))
			if err != nil {
				return err
			}
			return s.Listen()
		},
	}

	c.Flags().StringVarP(&filename, "config", "c", "", "Configuration file")
	return c
}
