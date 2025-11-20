package core

import "github.com/Masterminds/semver"

// ICore 核心接口
type ICore interface {
	// 版本操作
	ParseVersion(version string) (*semver.Version, error)
	SortVersions(versions []string) ([]string, error)
	Download(url string, version string, dst string) (string, error)
	Extract(src string, dst string) error

	// 缓存
	LoadCache() ([]string, error)
	SaveCache(versions []string) error
}
