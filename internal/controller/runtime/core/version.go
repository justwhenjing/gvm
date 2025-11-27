package core

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"github.com/justwhenjing/gokit/util/fileop"

	"github.com/justwhenjing/gvm/internal/util/httpcli"
)

// Download 下载版本
func (c *Core) Download(repo string, version string, destFolder string) (string, error) {
	// 1) 创建父目录
	tarName := fullName(version)
	if err := os.MkdirAll(destFolder, 0755); err != nil {
		return "", err
	}

	// 2) 下载版本
	downloadURL, _ := url.JoinPath(repo, tarName)
	dest := filepath.Join(destFolder, tarName)
	c.logger.Debug("download", "src", downloadURL, "dest", dest)

	client := httpcli.NewClient(
		httpcli.WithDebug(c.o.verbose),
	)

	// 3) 下载文件
	resp, err := client.GetWithOutput(downloadURL, dest)
	if err != nil {
		return "", err
	}

	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("%s returned status code %d", downloadURL, resp.StatusCode())
	}

	c.logger.Info("download completed")
	return dest, nil
}

// Extract 解压版本
func (c *Core) Extract(src string, dst string) error {
	c.logger.Debug("extract version", "src", src, "dst", dst)

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	return fileop.Extract(src, dst)
}

// fullName 完整的文件名
func fullName(version string) string {
	return fmt.Sprintf("go%s.%s-%s%s", version, runtime.GOOS, runtime.GOARCH, TarExt)
}
