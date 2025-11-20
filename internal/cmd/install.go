package cmd

import (
	"github.com/spf13/cobra"

	"github.com/justwhenjing/gvm/internal/controller/config"
	"github.com/justwhenjing/gvm/internal/controller/runtime"
	"github.com/justwhenjing/gvm/internal/util/log"
)

func NewInstallCmd(logger log.ILog, c *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "install [version]",
		Long: "install spec go version",
		RunE: func(cmd *cobra.Command, args []string) error {
			var version string
			if len(args) > 0 {
				version = args[0]
			}

			r := runtime.NewRuntime(logger, c)
			if err := r.Install(version); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
