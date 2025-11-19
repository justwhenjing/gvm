package cmd

import (
	"fmt"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

var (
	Version = "v1.0.0"
	Commit  = "none"
	date    = time.Now().Format(time.DateTime)
)

func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "version",
		Short:   "version",
		Example: "version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%-14s%s\n", "Version:", Version)
			fmt.Printf("%-14s%s\n", "Git Commit:", Commit)
			fmt.Printf("%-14s%s\n", "Build time:", date)
			fmt.Printf("%-14s%s\n", "Go version:", runtime.Version())
			fmt.Printf("%-14s%s/%s\n", "OS/Arch:", runtime.GOOS, runtime.GOARCH)
		},
	}

	return cmd
}
