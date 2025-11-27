package httpcli

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/schollz/progressbar/v3"
)

var _ IHttp = (*Client)(nil)

type Client struct {
	o *Option
}

func NewClient(opts ...OptionFunc) *Client {
	o := NewOption()
	o.ApplyOptions(opts...)
	return &Client{o: o}
}

func (c *Client) Get(url string, query map[string]string) (*resty.Response, error) {
	return c.o.client.R().SetQueryParams(query).Get(url)
}

func (c *Client) GetWithOutput(url string, output string) (*resty.Response, error) {
	// 1) 查看文件大小
	head, err := c.o.client.R().Head(url)
	if err != nil {
		return nil, err
	}
	var total int64
	contentLength := head.Header().Get("Content-Length")
	if contentLength != "" {
		total, _ = strconv.ParseInt(contentLength, 10, 64)
	}

	// 2) 创建进度条
	bar := progressbar.NewOptions64(
		total,
		progressbar.OptionSetDescription("Downloading"),
		progressbar.OptionSetWriter(os.Stdout),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowTotalBytes(true),
		progressbar.OptionSetWidth(20),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n") // 完成后强制换行
		}),
	)
	defer func() { _ = bar.Close() }()

	// 3) 开启进度
	done := make(chan struct{})

	ticker := time.NewTicker(200 * time.Millisecond) // 每200ms检查一次
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-done:
				current := currentSize(output)
				if total > 0 && current < total {
					current = total
				}
				_ = bar.Set64(current)
				return
			case <-ticker.C:
				current := currentSize(output)
				_ = bar.Set64(current)
			}
		}
	}()

	resp, err := c.o.client.R().SetOutput(output).Get(url)
	if err != nil {
		close(done)
		return nil, err
	}

	close(done)
	time.Sleep(200 * time.Millisecond)

	return resp, nil
}

func (c *Client) Post(url string, body interface{}) (*resty.Response, error) {
	return c.o.client.R().SetBody(body).Post(url)
}

func (c *Client) Patch(url string, body interface{}) (*resty.Response, error) {
	return c.o.client.R().SetBody(body).Patch(url)
}

func currentSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}
