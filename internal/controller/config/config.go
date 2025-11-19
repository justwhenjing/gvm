package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
)

// TODO 优化全局配置
const (
	DefaultRootDir  = "gvm"
	DefaultRepo     = "https://go.dev/dl/"
	DefaultCacheTTL = time.Duration(10) * time.Minute
	DefaultTagURL   = "https://raw.githubusercontent.com/kevincobain2000/gobrew/json/golang-tags.json"
)

type Config struct {
	RootDir    string        `json:"root_dir" validate:"required"`     // gvm根目录
	Repo       string        `json:"repo" validate:"required"`         // 版本仓库
	TagURL     string        `json:"tag_url" validate:"required"`      // 版本标签URL
	Verbose    bool          `json:"verbose" validate:"omitempty"`     // 是否显示详细信息
	Remote     bool          `json:"remote" validate:"omitempty"`      // 是否显示远程版本信息
	ClearCache bool          `json:"clear_cache" validate:"omitempty"` // 是否清理缓存
	CacheTTL   time.Duration `json:"cache_ttl" validate:"omitempty"`   // 缓存过期时间
}

func (c *Config) BackFill() error {
	// 根据环境变量设置,否则设置到~/gvm下
	if c.RootDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		c.RootDir = filepath.Join(home, DefaultRootDir)
	}

	return nil
}

func (c *Config) Validate() error {
	err := validator.New().Struct(c)
	if err == nil {
		return nil
	}

	errInfos, ok := err.(validator.ValidationErrors)
	if !ok {
		return fmt.Errorf("err to ValidationErrors failed")
	}

	var format string
	for _, errInfo := range errInfos {
		format = fmt.Sprintf("validate '%s'(tag '%s') failed:", errInfo.StructNamespace(), errInfo.ActualTag())
		if errInfo.Param() != "" {
			format = fmt.Sprintf("%s expect is '%v'", format, errInfo.Param())
		}
		format = fmt.Sprintf("%s, actual is %v", format, errInfo.Value())
	}
	return fmt.Errorf("%s", format)
}

func (c *Config) String() string {
	str, _ := json.Marshal(c)
	return string(str)
}
