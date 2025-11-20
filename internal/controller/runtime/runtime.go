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
		repoURL:       c.Repo,
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

// Use 使用指定版本
func (r *Runtime) Use(version string) error {
	if r.CurrentVersion() == version {
		r.logger.Info("already using", "version", version)
		return nil
	}

	// 设置go目录软链接
	_ = os.RemoveAll(r.o.currentGoDir)
	goDir := filepath.Join(r.o.versionsDir, version, "go")
	if err := os.MkdirAll(filepath.Dir(r.o.currentGoDir), 0755); err != nil {
		return err
	}
	if err := os.Symlink(goDir, r.o.currentGoDir); err != nil {
		return err
	}

	// 设置bin目录软链接
	_ = os.RemoveAll(r.o.currentBinDir)
	binDir := filepath.Join(r.o.versionsDir, version, "go", "bin")
	if err := os.MkdirAll(filepath.Dir(r.o.currentBinDir), 0755); err != nil {
		return err
	}
	if err := os.Symlink(binDir, r.o.currentBinDir); err != nil {
		return err
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
