package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/justwhenjing/gvm/internal/controller/config"
	"github.com/justwhenjing/gvm/internal/util/log"
)

func NewRootCmd() (*cobra.Command, error) {
	// 初始化logger(默认使用info)
	logger, err := log.NewLogger(
		os.Stdout,
		log.WithFormat(log.FormatCustom),
		log.WithShowLevel(true),
		log.WithColorful(false),
	)
	if err != nil {
		return nil, err
	}
	if err := logger.SetLevel(log.LevelInfo); err != nil {
		return nil, err
	}

	// 初始化配置
	c := &config.Config{}

	cmd := &cobra.Command{
		Use:               "gvm",
		Long:              "gvm tool is a tool for managing Go versions",
		SilenceUsage:      true,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Name() == "version" {
				return nil
			}

			if err := c.BackFill(); err != nil {
				return err
			}

			if err := c.Validate(); err != nil {
				return err
			}

			if c.Verbose {
				if err := logger.SetLevel(log.LevelDebug); err != nil {
					return err
				}
			}

			logger.Debug("show config", "config", c.String())

			return nil
		},
	}

	cmd.AddCommand(
		NewListCmd(logger, c),
		NewVersionCmd(),
	)

	// 设置选项
	cmd.PersistentFlags().StringVarP(&c.RootDir, "root", "", os.Getenv("GVM_ROOT"), "gvm root directory")
	cmd.PersistentFlags().StringVarP(&c.Repo, "repo", "", config.DefaultRepo, "gvm version repository")
	cmd.PersistentFlags().BoolVarP(&c.Verbose, "verbose", "v", false, "if show details")

	return cmd, nil
}
