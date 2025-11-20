package core

import "time"

// TODO 如何优化一下
type Option struct {
	cacheFile string        // 缓存文件
	ttl       time.Duration // 缓存过期时间
	verbose   bool          // 是否显示详细信息
}

func (o *Option) Apply(opts []OptionFunc) {
	for _, opt := range opts {
		opt(o)
	}
}

// 选项
type OptionFunc func(o *Option)
