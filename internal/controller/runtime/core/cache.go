package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Cache struct {
	Timestamp string   `json:"timestamp"`
	Versions  []string `json:"versions"`
}

func (c *Core) LoadCache() ([]string, error) {
	if c.o.cacheFile == "" {
		c.logger.Debug("cache file is not set")
		return nil, nil
	}

	fObj, err := os.Open(c.o.cacheFile)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = fObj.Close()
	}()

	decoder := json.NewDecoder(fObj)
	cache := &Cache{}
	if err := decoder.Decode(cache); err != nil {
		return nil, err
	}

	// 检查缓存是否过期
	timestamp, e := time.Parse(time.RFC3339, cache.Timestamp)
	if e != nil {
		return nil, err
	}

	if time.Now().UTC().After(timestamp.Add(c.o.ttl)) {
		return nil, fmt.Errorf("cache expired")
	}

	return cache.Versions, nil
}

func (c *Core) SaveCache(versions []string) error {
	// 外部调用设置缓存清理
	if c.o.cacheFile == "" {
		c.logger.Debug("cache file is not set")
		return nil
	}

	// 覆盖写入缓存文件
	if err := os.MkdirAll(filepath.Dir(c.o.cacheFile), 0755); err != nil {
		return err
	}
	fObj, err := os.OpenFile(c.o.cacheFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = fObj.Close()
	}()

	encoder := json.NewEncoder(fObj)
	encoder.SetIndent("", "  ")
	cache := &Cache{
		Timestamp: time.Now().Format(time.RFC3339),
		Versions:  versions,
	}
	return encoder.Encode(cache)
}
