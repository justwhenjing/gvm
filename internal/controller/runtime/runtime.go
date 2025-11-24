package runtime

import (
	"fmt"
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
		currentDir:   filepath.Join(c.RootDir, "current"),
		versionsDir:  filepath.Join(c.RootDir, "versions"),
		downloadsDir: filepath.Join(c.RootDir, "downloads"),
		goPathDir:    filepath.Join(c.RootDir, "workspace"),
		repoURL:      c.Repo,
		tagURL:       c.TagURL,
		verbose:      c.Verbose,
		remote:       c.Remote,
	}
	o.Apply(opts)

	return &Runtime{
		logger: logger,
		core:   core.NewCore(logger.With("runtime", "core"), c),
		o:      o,
	}
}

// Use 使用指定版本
func (r *Runtime) Use(version string) error {
	if r.CurrentVersion() == version {
		r.logger.Info("already using", "version", version)
		return nil
	}

	// 设置当前版本软链接
	_ = os.RemoveAll(r.o.currentDir)
	goDir := filepath.Join(r.o.versionsDir, version, "go")
	if err := os.MkdirAll(filepath.Dir(r.o.currentDir), 0755); err != nil {
		return err
	}
	if err := os.Symlink(goDir, r.o.currentDir); err != nil {
		return err
	}

	// 创建GOPATH目录
	if err := os.MkdirAll(r.o.goPathDir, 0755); err != nil {
		return err
	}

	// 判断环境变量
	if err := r.CheckEnv(); err != nil {
		r.logger.Warn("check env failed", "error", err)
	}

	r.logger.Info("using", "version", version)
	return nil
}

// Install 安装指定版本
func (r *Runtime) Install(version string) error {
	if version == "" {
		// 不指定版本则获取最新的稳定版本
		latestVersion, err := r.LatestRemoteVersion()
		if err != nil {
			return err
		}
		version = latestVersion
	}
	r.logger.Info("installing", "version", version)

	// 查看版本是否已存在
	if r.ExistVersion(version) {
		r.logger.Info("version already exists", "version", version)
		if r.CurrentVersion() != version {
			return r.Use(version)
		}
		return nil
	}

	// 下载版本
	_ = os.RemoveAll(r.o.downloadsDir)
	defer func() {
		_ = os.RemoveAll(r.o.downloadsDir)
	}()
	tarName, err := r.core.Download(r.o.repoURL, version, r.o.downloadsDir)
	if err != nil {
		return err
	}

	// 解压版本
	dst := filepath.Join(r.o.versionsDir, version)
	if err := r.core.Extract(tarName, dst); err != nil {
		_ = os.RemoveAll(dst)
		return err
	}

	// 安装版本
	if err := r.Use(version); err != nil {
		_ = os.RemoveAll(dst)
		return err
	}

	return nil
}

// Uninstall 卸载指定版本
func (r *Runtime) Uninstall(version string) error {
	if version == "" {
		return fmt.Errorf("version is required")
	}

	if r.CurrentVersion() == version {
		return fmt.Errorf("cannot uninstall current version")
	}

	// 清理versions目录
	r.logger.Info("uninstalling", "version", version)

	versionDir := filepath.Join(r.o.versionsDir, version)
	_ = os.RemoveAll(versionDir)

	r.logger.Info("uninstall completed")
	return nil
}

// Prune 清理所有版本
func (r *Runtime) Prune() error {
	entries, err := os.ReadDir(r.o.versionsDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		r.logger.Info("pruning", "version", entry.Name())
		if r.CurrentVersion() == entry.Name() {
			_ = os.RemoveAll(r.o.currentDir)
		}
		_ = os.RemoveAll(filepath.Join(r.o.versionsDir, entry.Name()))
	}

	r.logger.Info("prune completed")
	return nil
}

// List 列举版本
func (r *Runtime) List(filter string) error {
	if r.o.remote {
		// 远程版本列举
		// 1) 优先从缓存中加载版本
		versions, err := r.core.LoadCache()
		if err != nil {
			r.logger.Debug("load cache failed", "error", err.Error())
		}

		// 2) 从远程仓库获取版本
		if len(versions) == 0 {
			versions, err = r.RemoteVersions()
			if err != nil {
				return err
			}

			// 保存缓存
			if err := r.core.SaveCache(versions); err != nil {
				return err
			}
			r.logger.Debug("save cache", "versions", versions)
		}

		// 3) 版本分组
		keys, group, err := r.GroupVersions(versions)
		if err != nil {
			return err
		}
		r.logger.Debug("group versions", "keys", keys, "group", group)

		for _, key := range keys {
			if filter == "" {
				r.logger.Info(key, "versions", group[key])
				continue
			}

			if strings.Contains(filter, key) {
				r.logger.Info(key, "versions", group[key])
			}
		}
		return nil
	}

	// 本地版本列举
	entries, err := os.ReadDir(r.o.versionsDir)
	if err != nil {
		return err
	}

	versions := make([]string, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return err
		}
		versions = append(versions, info.Name())
	}

	// 标记当前版本
	cv := r.CurrentVersion()
	sortedVersions, err := r.core.SortVersions(versions)
	if err != nil {
		return err
	}

	for _, version := range sortedVersions {
		if version == cv {
			r.logger.Info(version + " *")
		} else {
			r.logger.Info(version)
		}
	}
	r.logger.Info("")

	if cv != "" {
		r.logger.Info("current", "version", cv)
	}

	return nil
}

// CheckEnv 检查环境变量
func (r *Runtime) CheckEnv() error {
	if os.Getenv("PATH") == "" {
		return fmt.Errorf("PATH not set, please check")
	}

	// 检查GOROOT环境变量
	if os.Getenv("GOROOT") != r.o.currentDir {
		return fmt.Errorf("GOROOT set wrong, please set %s as GOROOT", r.o.currentDir)
	}
	binDir := filepath.Join(r.o.currentDir, "bin")
	if !strings.Contains(os.Getenv("PATH"), binDir) {
		return fmt.Errorf("PATH set wrong, please add %s to PATH", binDir)
	}

	// 检查GOPATH环境变量
	if os.Getenv("GOPATH") != r.o.goPathDir {
		return fmt.Errorf("GOPATH set wrong, please set %s as GOPATH", r.o.goPathDir)
	}
	binDir = filepath.Join(r.o.goPathDir, "bin")
	if !strings.Contains(os.Getenv("PATH"), binDir) {
		return fmt.Errorf("PATH set wrong, please add %s to PATH", binDir)
	}

	return nil
}
