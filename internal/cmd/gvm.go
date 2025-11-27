package cmd

import (
	"os"

	"github.com/justwhenjing/gokit/infra/log"
	"github.com/spf13/cobra"

	"github.com/justwhenjing/gvm/internal/controller/config"
)

func NewRootCmd() (*cobra.Command, error) {
	// 初始化logger(默认使用info)
	logger, err := log.NewLogger(
		log.WithFormat(log.FormatJSON),
		log.WithRemoveTime(true),
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

			// 配置回填
			if err := c.BackFill(); err != nil {
				return err
			}

			// 配置检查
			if err := c.Validate(); err != nil {
				return err
			}

			// 日志级别设置
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
		NewInstallCmd(logger, c),
		NewUninstallCmd(logger, c),
		NewUseCmd(logger, c),
		NewVersionCmd(),
	)

	// 设置选项
	cmd.PersistentFlags().StringVarP(&c.RootDir, "root", "", os.Getenv("GVM_HOME"), "gvm root directory")
	cmd.PersistentFlags().StringVarP(&c.Repo, "repo", "", config.DefaultRepo, "gvm version repository")
	cmd.PersistentFlags().BoolVarP(&c.Verbose, "verbose", "v", false, "if show details")

	return cmd, nil
}
