package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/justwhenjing/gvm/internal/controller/config"
	"github.com/justwhenjing/gvm/internal/controller/runtime"
	"github.com/justwhenjing/gvm/internal/util/log"
)

func NewListCmd(logger log.ILog, c *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "list [filter]",
		Long: "list go versions",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if c.ClearCache {
				// 清理缓存文件
				_ = os.RemoveAll(filepath.Join(c.RootDir, "cache.json"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var filter string
			if len(args) > 0 {
				filter = args[0]
			}

			r := runtime.NewRuntime(logger, c)
			if err := r.ListVersions(filter); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.PersistentFlags().BoolVarP(&c.Remote, "remote", "", false, "if show remote versions")
	cmd.PersistentFlags().BoolVarP(&c.ClearCache, "clear-cache", "", false, "if clear cache")
	cmd.PersistentFlags().StringVarP(&c.TagURL, "tag-url", "", config.DefaultTagURL, "tag url")
	cmd.PersistentFlags().DurationVarP(&c.CacheTTL, "cache-ttl", "", config.DefaultCacheTTL, "cache ttl")

	return cmd
}
