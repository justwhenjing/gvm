package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/justwhenjing/gokit/infra/log"
	"github.com/justwhenjing/gvm/internal/controller/config"
	"github.com/justwhenjing/gvm/internal/controller/runtime"
)

func NewUninstallCmd(logger log.ILog, c *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "uninstall [version]",
		Long: "uninstall spec go version",
		RunE: func(cmd *cobra.Command, args []string) error {
			r := runtime.NewRuntime(logger, c)
			if c.Prune {
				return r.Prune()
			}

			if len(args) < 1 {
				return fmt.Errorf("version is required")
			}
			return r.Uninstall(args[0])
		},
	}

	cmd.PersistentFlags().BoolVarP(&c.Prune, "prune", "", false, "if prune all versions")
	return cmd
}
