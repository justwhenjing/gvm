package runtime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver"

	"github.com/justwhenjing/gvm/internal/controller/runtime/core"
	"github.com/justwhenjing/gvm/internal/util/httpcli"
)

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

// LatestRemoteVersion 获取最新远程版本
func (r *Runtime) LatestRemoteVersion() (string, error) {
	remoteVersions, err := r.RemoteVersions()
	if err != nil {
		return "", err
	}

	stableVersions := make([]*semver.Version, 0)
	for _, version := range remoteVersions {
		// 判断是否支持
		if core.NotSupportedVersion(version) {
			continue
		}

		// 判断是否为beat版本
		if core.IsBetaOrRC(version) {
			continue
		}

		v, err := semver.NewVersion(version)
		if err != nil {
			r.logger.Debug("parse version %s failed: %w", version, err)
			continue
		}
		stableVersions = append(stableVersions, v)
	}

	if len(stableVersions) == 0 {
		return "", fmt.Errorf("no stable version found in remote versions")
	}

	sort.Sort(semver.Collection(stableVersions))

	return core.FormatVersion(
		stableVersions[len(stableVersions)-1].String(),
	), nil
}

type Tag struct {
	Ref string `json:"ref"`
}

// RemoteVersions 获取远程版本
func (r *Runtime) RemoteVersions() ([]string, error) {
	client := httpcli.NewClient(httpcli.WithDebug(r.o.verbose))
	response, err := client.Get(r.o.tagURL, nil)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("get remote versions failed, status code: %d", response.StatusCode())
	}

	tags := make([]Tag, 0)
	if err := json.Unmarshal(response.Body(), &tags); err != nil {
		return nil, err
	}

	// 去除refs/tags/ 前缀和 go后缀
	versions := make([]string, 0)
	for _, tag := range tags {
		ref := strings.ReplaceAll(tag.Ref, "refs/tags/", "")
		if strings.HasPrefix(ref, "go") {
			versions = append(versions, strings.TrimPrefix(ref, "go"))
		}
	}
	return versions, nil
}

// GroupVersions 版本分组
func (r *Runtime) GroupVersions(versions []string) ([]string, map[string][]string, error) {
	// 1) 分组
	group := make(map[string][]string)
	for _, version := range versions {
		// 分割版本号
		parts := strings.Split(version, ".")
		if len(parts) <= 1 {
			continue
		}

		// 主版本号
		majorVersion := fmt.Sprintf("%s.%s", parts[0], parts[1])
		// 带有beta/rc版本号的,直接去掉beta/rc后缀
		matches := core.MatchBetaOrRC(majorVersion)
		if len(matches) >= 1 {
			majorVersion = strings.Split(version, matches[0])[0]
		}

		// 过滤不支持的版本
		if core.NotSupportedVersion(majorVersion) {
			continue
		}
		group[majorVersion] = append(group[majorVersion], version)
	}

	// 2) 排序(使用语义化版本)
	versionsSemantic := make([]*semver.Version, 0)
	for key := range group {
		v, err := semver.NewVersion(key)
		if err != nil {
			return nil, nil, fmt.Errorf("parse version %s failed: %w", key, err)
		}
		versionsSemantic = append(versionsSemantic, v)
	}
	sort.Sort(semver.Collection(versionsSemantic))
	keys := make([]string, 0, len(versionsSemantic))
	for _, v := range versionsSemantic {
		// 语义化版本会自动增加.0,需要去掉
		keys = append(keys, strings.TrimSuffix(v.String(), ".0"))
	}

	return keys, group, nil
}

// ExistVersion 已存在版本
func (r *Runtime) ExistVersion(version string) bool {
	versionDir := filepath.Join(r.o.versionsDir, version, "go")
	_, err := os.Stat(versionDir)
	return err == nil
}
