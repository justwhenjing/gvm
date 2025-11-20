package core

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/c4milo/unpackit"
	"github.com/schollz/progressbar/v3"

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

	// 3) 配置进度条
	done := make(chan bool)

	headResp, err := client.Head(downloadURL)
	if err != nil {
		return "", err
	}
	var total int64
	contentLength := headResp.Header().Get("Content-Length")
	if contentLength != "" {
		total, _ = strconv.ParseInt(contentLength, 10, 64)
	}

	bar := progressbar.DefaultBytes(total, "Downloading")
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				info, err := os.Stat(dest)
				if err == nil {
					current := info.Size()
					_ = bar.Set64(current)
				}
			case <-done:
				// 设置最终进度为 100%
				if total > 0 {
					_ = bar.Set64(total)
				}
				return
			}
		}
	}()

	resp, err := client.GetWithOutput(downloadURL, dest)
	if err != nil {
		close(done)
		return "", err
	}

	if resp.StatusCode() != http.StatusOK {
		close(done)
		return "", fmt.Errorf("%s returned status code %d", downloadURL, resp.StatusCode())
	}

	close(done)
	// 等待完成
	time.Sleep(300 * time.Millisecond)

	c.logger.Info("download completed")
	return dest, nil
}

// Extract 解压版本
func (c *Core) Extract(src string, dst string) error {
	c.logger.Debug("extract version", "src", src, "dst", dst)

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// #nosec G304
	fObj, err := os.Open(src)
	if err != nil {
		return err
	}

	// 支持解压 tar.gz zip
	return unpackit.Unpack(fObj, dst)
}

// fullName 完整的文件名
func fullName(version string) string {
	return fmt.Sprintf("go%s.%s-%s%s", version, runtime.GOOS, runtime.GOARCH, TarExt)
}
