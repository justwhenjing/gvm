package cmd

import (
	"github.com/spf13/cobra"
)

func NewUseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use",
		Short: "use",
	}

	return cmd
}
