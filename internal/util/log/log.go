package log

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
)

var _ ILog = (*Logger)(nil)

type Logger struct {
	leveler *slog.LevelVar // 等级管理器(debug < info < warn < error)
	logger  *slog.Logger   // 日志接口
}

func NewLogger(w io.Writer, opts ...OptionFunc) (*Logger, error) {
	o := &Option{
		Format: FormatText,
	}
	o.Apply(opts...)

	leveler := &slog.LevelVar{}
	var handler slog.Handler

	// 支持不同格式的日志输出
	switch o.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: leveler,
		})
	case FormatText:
		handler = slog.NewTextHandler(w, &slog.HandlerOptions{
			Level: leveler,
		})
	case FormatCustom:
		handler = NewCustomHandler(w, o, &slog.HandlerOptions{
			Level: leveler,
		})
	default:
		return nil, fmt.Errorf("invalid format: %s", o.Format)
	}

	return &Logger{
		leveler: leveler,
		logger:  slog.New(handler),
	}, nil
}

func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

func (l *Logger) SetLevel(level Level) error {
	var slogLevel slog.Level

	switch strings.ToUpper(string(level)) {
	case string(LevelDebug):
		slogLevel = slog.LevelDebug
	case string(LevelInfo):
		slogLevel = slog.LevelInfo
	case string(LevelWarn):
		slogLevel = slog.LevelWarn
	case string(LevelError):
		slogLevel = slog.LevelError
	default:
		return fmt.Errorf("invalid level: %s", level)
	}

	l.leveler.Set(slogLevel)
	return nil
}

func (l *Logger) With(args ...any) ILog {
	l.logger = l.logger.With(args...)
	return l
}
