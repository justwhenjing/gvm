package runtime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/justwhenjing/gvm/internal/util/httpcli"
)

type Tag struct {
	Ref string `json:"ref"`
}

func (r *Runtime) ListVersions(filter string) error {
	if r.o.remote {
		return r.ListRemoteVersions(filter)
	}
	return r.ListLocalVersions()
}

func (r *Runtime) ListRemoteVersions(filter string) error {
	// 1) 优先从缓存中加载版本
	versions := make([]string, 0)
	cache, err := r.core.LoadCache()
	if err != nil {
		r.logger.Debug("load cache failed", "error", err.Error())
	} else {
		versions = cache
	}

	// 2) 从远程仓库获取版本
	if len(versions) == 0 {
		client := httpcli.NewClient(httpcli.WithDebug(r.o.verbose))
		response, err := client.Get(r.o.tagURL, nil)
		if err != nil {
			return err
		}
		if response.StatusCode() != http.StatusOK {
			return fmt.Errorf("get remote versions failed, status code: %d", response.StatusCode())
		}

		tags := make([]Tag, 0)
		if err := json.Unmarshal(response.Body(), &tags); err != nil {
			return err
		}
		// 去除refs/tags/ 前缀和 go后缀
		for _, tag := range tags {
			ref := strings.ReplaceAll(tag.Ref, "refs/tags/", "")
			if strings.HasPrefix(ref, "go") {
				versions = append(versions, strings.TrimPrefix(ref, "go"))
			}
		}

		// 保存缓存
		if err := r.core.SaveCache(versions); err != nil {
			return err
		}
		r.logger.Debug("save cache", "versions", versions)
	}

	// 3) 版本分组
	keys, group, err := r.groupVersions(versions)
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

// groupVersions 版本分组
func (r *Runtime) groupVersions(versions []string) ([]string, map[string][]string, error) {
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
		r := regexp.MustCompile("beta.*|rc.*")
		matches := r.FindAllString(majorVersion, -1)
		if len(matches) == 1 {
			majorVersion = strings.Split(version, matches[0])[0]
		}

		// 过滤不支持的版本
		if notSupportedVersion(majorVersion) {
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

// notSupportedVersion 不支持的版本
func notSupportedVersion(version string) bool {
	blackListVersions := []string{"1.0", "1.1", "1.2", "1.3", "1.4"}
	for _, v := range blackListVersions {
		if version == v {
			return true
		}
	}
	return false
}

func (r *Runtime) ListLocalVersions() error {
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
