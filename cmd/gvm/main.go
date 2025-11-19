package main

import (
	"os"

	"github.com/justwhenjing/gvm/internal/cmd"
)

func main() {
	rootCmd, err := cmd.NewRootCmd()
	if err != nil {
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
