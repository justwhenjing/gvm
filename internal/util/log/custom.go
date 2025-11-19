package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"
)

type CustomHandler struct {
	writer io.Writer

	opts        *Option
	handlerOpts *slog.HandlerOptions
}

func NewCustomHandler(w io.Writer, opts *Option, handlerOpts *slog.HandlerOptions) slog.Handler {
	if handlerOpts == nil {
		handlerOpts = &slog.HandlerOptions{}
	}
	return &CustomHandler{
		opts:        opts,
		handlerOpts: handlerOpts,
		writer:      w,
	}
}

func (h *CustomHandler) Enabled(ctx context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.handlerOpts.Level != nil {
		minLevel = h.handlerOpts.Level.Level()
	}
	return level >= minLevel
}

func (h *CustomHandler) Handle(ctx context.Context, record slog.Record) error {
	var parts []string

	// 1) 时间戳 (可选)
	if h.opts.showTimestamp {
		timeStr := record.Time.Format(time.DateTime)
		if h.opts.colorful {
			timeStr = grayColor(timeStr)
		}
		parts = append(parts, timeStr)
	}

	// 2) 日志级别 (带颜色和图标)
	if h.opts.showLevel {
		levelStr := h.formatLevel(record.Level)
		parts = append(parts, levelStr)
	}

	// 3) 消息内容
	message := record.Message
	if h.opts.colorful {
		message = getMessageColor(record.Level)(message)
	}
	parts = append(parts, message)

	// 4) 附加属性 (如果有)
	if record.NumAttrs() > 0 {
		attrsStr := h.formatAttrs(record)
		if attrsStr != "" {
			if h.opts.colorful {
				attrsStr = grayColor(attrsStr)
			}
			parts = append(parts, attrsStr)
		}
	}

	// 5) 组合所有部分
	output := strings.Join(parts, " ") + "\n"
	_, err := h.writer.Write([]byte(output))
	return err
}

// formatLevel 格式化日志级别
func (h *CustomHandler) formatLevel(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		if h.opts.colorful {
			return blueColor(fmt.Sprintf("[%s]", level.String()))
		}
		return fmt.Sprintf("[%s]", level.String())
	case slog.LevelInfo:
		if h.opts.colorful {
			return greenColor(fmt.Sprintf("[%s]", level.String()))
		}
		return fmt.Sprintf("[%s]", level.String())
	case slog.LevelWarn:
		if h.opts.colorful {
			return yellowColor(fmt.Sprintf("[%s]", level.String()))
		}
		return fmt.Sprintf("[%s]", level.String())
	case slog.LevelError:
		if h.opts.colorful {
			return redColor(fmt.Sprintf("[%s]", level.String()))
		}
		return fmt.Sprintf("[%s]", level.String())
	default:
		if h.opts.colorful {
			return cyanColor(fmt.Sprintf("[%s]", level.String()))
		}
		return fmt.Sprintf("[%s]", level.String())
	}
}

// formatAttrs 格式化附加属性
func (h *CustomHandler) formatAttrs(record slog.Record) string {
	var attrs []string
	record.Attrs(func(attr slog.Attr) bool {
		// 跳过已经显示的标准属性
		if attr.Key == "time" || attr.Key == "level" || attr.Key == "msg" {
			return true
		}
		if h.opts.colorful {
			attrs = append(attrs, fmt.Sprintf("%s=%v", cyanColor(attr.Key), attr.Value.Any()))
		} else {
			attrs = append(attrs, fmt.Sprintf("%s=%v", attr.Key, attr.Value.Any()))
		}
		return true
	})
	return strings.Join(attrs, " ")
}

func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CustomHandler{
		opts:        h.opts,
		handlerOpts: h.handlerOpts,
		writer:      h.writer,
	}
}

func (h *CustomHandler) WithGroup(name string) slog.Handler {
	return &CustomHandler{
		opts:        h.opts,
		handlerOpts: h.handlerOpts,
		writer:      h.writer,
	}
}
