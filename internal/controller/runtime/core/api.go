package core

import "github.com/Masterminds/semver"

// ICore 核心接口
type ICore interface {
	ParseVersion(version string) (*semver.Version, error)
	SortVersions(versions []string) ([]string, error)
	LoadCache() ([]string, error)
	SaveCache(versions []string) error
}
