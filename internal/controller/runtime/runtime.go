package runtime

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/justwhenjing/gvm/internal/controller/config"
	"github.com/justwhenjing/gvm/internal/controller/runtime/core"
	"github.com/justwhenjing/gvm/internal/util/log"
)

type Runtime struct {
	logger log.ILog   // 日志接口
	core   core.ICore // 核心接口

	o *Option // 选项
}

func NewRuntime(logger log.ILog, c *config.Config, opts ...OptionFunc) IRuntime {
	o := &Option{
		currentDir:    filepath.Join(c.RootDir, "current"),
		currentBinDir: filepath.Join(c.RootDir, "current", "bin"),
		currentGoDir:  filepath.Join(c.RootDir, "current", "go"),
		versionsDir:   filepath.Join(c.RootDir, "versions"),
		downloadsDir:  filepath.Join(c.RootDir, "downloads"),
		tagURL:        c.TagURL,
		verbose:       c.Verbose,
		remote:        c.Remote,
	}
	o.Apply(opts)

	return &Runtime{
		logger: logger,
		core:   core.NewCore(logger.With("runtime", "core"), c),
		o:      o,
	}
}

// CurrentVersion 查看当前版本
func (r *Runtime) CurrentVersion() string {
	// 查看软连接实际目录
	fp, err := os.Readlink(r.o.currentBinDir)
	if err != nil {
		r.logger.Debug("eval symlinks failed", "error", err)
		return core.NoneVersion
	}
	r.logger.Debug("current version", "actual_path", fp)

	// 去除go/bin后缀,对应的父目录名即为版本名
	version := strings.TrimSuffix(fp, filepath.Join("go", "bin"))
	version = filepath.Base(version)
	if version == "." {
		return core.NoneVersion
	}
	return version
}
