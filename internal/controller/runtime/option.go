package runtime

type Option struct {
	currentDir    string // 当前版本目录
	currentBinDir string // 当前版本二进制目录
	currentGoDir  string // 当前版本go目录
	versionsDir   string // 版本目录
	downloadsDir  string // 下载目录
	tagURL        string // 版本标签URL
	verbose       bool   // 是否显示详细信息
	remote        bool   // 是否显示远程版本信息
}

func (o *Option) Apply(opts []OptionFunc) {
	for _, opt := range opts {
		opt(o)
	}
}

// 选项(后续可以使用)
type OptionFunc func(o *Option)
