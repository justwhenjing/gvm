package httpcli

import (
	"crypto/tls"
	"time"

	"github.com/go-resty/resty/v2"
)

type Option struct {
	client *resty.Client
}

func NewOption() *Option {
	return &Option{
		client: resty.New(),
	}
}

func (o *Option) ApplyOptions(opts ...OptionFunc) {
	for _, opt := range opts {
		opt(o)
	}
}

// 可选项
type OptionFunc func(*Option)

// WithHeader 设置请求头
func WithHeader(header map[string]string) OptionFunc {
	return func(o *Option) {
		for k, v := range header {
			o.client.SetHeader(k, v)
		}
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) OptionFunc {
	return func(o *Option) {
		o.client.SetTimeout(timeout)
	}
}

// WithInsecureSkipVerify 设置是否跳过证书验证
func WithInsecureSkipVerify(insecure bool) OptionFunc {
	return func(o *Option) {
		o.client.SetTLSClientConfig(&tls.Config{
			InsecureSkipVerify: insecure,
		})
	}
}

// WithDebug 设置是否开启调试模式
func WithDebug(debug bool) OptionFunc {
	return func(o *Option) {
		o.client.SetDebug(debug)
	}
}

// WithRetryCount 设置重试次数
func WithRetryCount(retry int) OptionFunc {
	return func(o *Option) {
		o.client.SetRetryCount(retry)
	}
}
