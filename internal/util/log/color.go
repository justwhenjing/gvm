package log

import "log/slog"

// ANSI 颜色代码
const (
	colorReset  = "\033[0m"  // 重置颜色
	colorRed    = "\033[31m" // 红色
	colorGreen  = "\033[32m" // 绿色
	colorYellow = "\033[33m" // 黄色
	colorBlue   = "\033[34m" // 蓝色
	colorCyan   = "\033[36m" // 青色
	colorGray   = "\033[90m" // 灰色
	colorWhite  = "\033[97m" // 白色
)

// 颜色包装函数
func colorize(colorCode, text string) string {
	return colorCode + text + colorReset
}

func redColor(text string) string    { return colorize(colorRed, text) }
func greenColor(text string) string  { return colorize(colorGreen, text) }
func yellowColor(text string) string { return colorize(colorYellow, text) }
func blueColor(text string) string   { return colorize(colorBlue, text) }
func cyanColor(text string) string   { return colorize(colorCyan, text) }
func grayColor(text string) string   { return colorize(colorGray, text) }
func whiteColor(text string) string  { return colorize(colorWhite, text) }

// 根据日志级别获取消息颜色
func getMessageColor(level slog.Level) func(string) string {
	switch level {
	case slog.LevelDebug:
		return grayColor
	case slog.LevelInfo:
		return whiteColor
	case slog.LevelWarn:
		return yellowColor
	case slog.LevelError:
		return redColor
	default:
		return whiteColor
	}
}
