package log

type Format string

const (
	FormatJSON   Format = "json"
	FormatText   Format = "text"
	FormatCustom Format = "custom" // 自定义格式输出
)

type Option struct {
	Format        Format
	showTimestamp bool // 显示时间戳
	showLevel     bool // 显示日志级别
	colorful      bool // 是否彩色输出
}

func DefaultOption() *Option {
	return &Option{
		Format: FormatText, // 默认采用text格式输出
	}
}

func (o *Option) Apply(opts ...OptionFunc) {
	for _, opt := range opts {
		opt(o)
	}
}

// 选项
type OptionFunc func(o *Option)

func WithFormat(format Format) OptionFunc {
	return func(o *Option) {
		o.Format = format
	}
}

func WithShowTimestamp(show bool) OptionFunc {
	return func(o *Option) {
		o.showTimestamp = show
	}
}

func WithShowLevel(show bool) OptionFunc {
	return func(o *Option) {
		o.showLevel = show
	}
}

func WithColorful(colorful bool) OptionFunc {
	return func(o *Option) {
		o.colorful = colorful
	}
}
