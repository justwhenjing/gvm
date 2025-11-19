package core

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver"

	"github.com/justwhenjing/gvm/internal/controller/config"
	"github.com/justwhenjing/gvm/internal/util/log"
)

const (
	NoneVersion = "None"
)

type Core struct {
	logger log.ILog // 日志接口

	o *Option // 选项
}

func NewCore(logger log.ILog, conf *config.Config, opts ...OptionFunc) ICore {
	o := &Option{
		cacheFile: filepath.Join(conf.RootDir, "cache.json"),
		ttl:       conf.CacheTTL,
	}
	o.Apply(opts)

	return &Core{
		logger: logger,
		o:      o,
	}
}

func (c *Core) ParseVersion(version string) (*semver.Version, error) {
	if version == "" || version == NoneVersion {
		return nil, fmt.Errorf("no version provided")
	}

	semverVersion, err := semver.NewVersion(formatVersion(version))
	if err != nil {
		return nil, fmt.Errorf("parse version %s failed: %w", version, err)
	}
	return semverVersion, nil
}

// SortVersions 排序版本号
func (c *Core) SortVersions(versions []string) ([]string, error) {
	result := make([]string, 0)
	semverVersions := make([]*semver.Version, 0)
	rcBetaVersions := make([]string, 0)

	// 分解beta/rc版本和语义化版本
	for _, version := range versions {
		if isBetaOrRC(version) {
			rcBetaVersions = append(rcBetaVersions, version)
		} else {
			// 解析为语义化版本
			v, err := semver.NewVersion(version)
			if err != nil {
				return nil, fmt.Errorf("semver version failed: %w", err)
			}
			semverVersions = append(semverVersions, v)
		}
	}

	// 处理语义化版本
	sort.Sort(semver.Collection(semverVersions))
	for _, versionSemantic := range semverVersions {
		version := versionSemantic.String()
		// For versions < 1.21.0, display as major.minor
		//
		if versionSemantic.Major() == 1 && versionSemantic.Minor() < 21 && versionSemantic.Patch() == 0 {
			version = fmt.Sprintf("%d.%d", versionSemantic.Major(), versionSemantic.Minor())
		}
		result = append(result, version)
	}

	// 处理beta/rc版本
	result = append(result, rcBetaVersions...)
	return result, nil
}

// formatVersion 格式化版本号
func formatVersion(version string) string {
	// Remove @latest and @dev-latest suffixes
	version = strings.TrimSuffix(version, "@latest")
	version = strings.TrimSuffix(version, "@dev-latest")

	// Remove .x and x suffixes
	version = strings.TrimSuffix(version, ".x")
	version = strings.TrimSuffix(version, "x")
	return version
}

// isBetaOrRC 判断是否是beta或rc版本
func isBetaOrRC(version string) bool {
	re := regexp.MustCompile("beta.*|rc.*")
	matches := re.FindAllString(version, -1)
	return len(matches) > 0
}
