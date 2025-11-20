package core

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/schollz/progressbar/v3"

	"github.com/justwhenjing/gvm/internal/controller/config"
	"github.com/justwhenjing/gvm/internal/util/httpcli"
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
		verbose:   conf.Verbose,
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

	semverVersion, err := semver.NewVersion(FormatVersion(version))
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
		if IsBetaOrRC(version) {
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

// FormatVersion 格式化版本号
func FormatVersion(version string) string {
	// Remove @latest and @dev-latest suffixes
	version = strings.TrimSuffix(version, "@latest")
	version = strings.TrimSuffix(version, "@dev-latest")

	// Remove .x and x suffixes
	version = strings.TrimSuffix(version, ".x")
	version = strings.TrimSuffix(version, "x")
	return version
}

// IsBetaOrRC 判断是否是beta或rc版本
func IsBetaOrRC(version string) bool {
	return len(MatchBetaOrRC(version)) > 0
}

// MatchBetaOrRC 匹配beta或rc版本
func MatchBetaOrRC(version string) []string {
	re := regexp.MustCompile("beta.*|rc.*")
	return re.FindAllString(version, -1)
}

// NotSupportedVersion 不支持的版本
func NotSupportedVersion(version string) bool {
	blackListVersions := []string{"1.0", "1.1", "1.2", "1.3", "1.4"}
	for _, v := range blackListVersions {
		if version == v {
			return true
		}
	}
	return false
}

// DownloadVersion 下载版本
func DownloadVersion(url string, tarName string, destFolder string) error {
	if err := os.MkdirAll(destFolder, 0755); err != nil {
		return err
	}

	client := httpcli.NewClient()
	response, err := client.Get(url, nil)
	if err != nil {
		return err
	}
	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("download version failed, status code: %d", response.StatusCode())
	}

	dest := filepath.Join(destFolder, tarName)
	fObj, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		_ = fObj.Close()
	}()

	bar := progressbar.DefaultBytes(
		response.Size(),
		"Downloading",
	)

	_, err = io.Copy(io.MultiWriter(fObj, bar), bytes.NewReader(response.Body()))
	if err != nil {
		return err
	}

	return nil
}
