package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/justwhenjing/gvm/internal/controller/config"
	"github.com/justwhenjing/gvm/internal/controller/runtime"
	"github.com/justwhenjing/gvm/internal/util/log"
)

func NewUseCmd(logger log.ILog, c *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "use [version]",
		Long: "use spec go version",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("version is required")
			}
			version := args[0]

			r := runtime.NewRuntime(logger, c)
			if err := r.Use(version); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
