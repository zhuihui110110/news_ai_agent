// @CopyRight (c) 2026
// @Author 　 : a123
// @CreateTime: 2026/10/23
// @Description: 全局日志(zap 精简封装, 参照 stock_ai xlog 用法: log.Init 后全局 Debugf/Errorw...)

package log

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *zap.Logger
	sugar        *zap.SugaredLogger
)

// Init 初始化全局日志: level 为 debug/info/warn/error, director 为日志目录
func Init(level, director string) error {
	if globalLogger != nil {
		return nil
	}
	if err := os.MkdirAll(director, 0o755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}
	lv := zapcore.InfoLevel
	_ = lv.UnmarshalText([]byte(level))

	encoder := zap.NewProductionEncoderConfig()
	encoder.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder.EncodeLevel = zapcore.CapitalLevelEncoder

	console := zapcore.AddSync(os.Stdout)
	file := zapcore.AddSync(newRollFile(director))

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewConsoleEncoder(encoder), console, lv),
		zapcore.NewCore(zapcore.NewJSONEncoder(encoder), file, lv),
	)
	globalLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar = globalLogger.Sugar()
	return nil
}

// Sync 刷盘
func Sync() {
	if globalLogger != nil {
		_ = globalLogger.Sync()
	}
}

func newRollFile(director string) zapcore.WriteSyncer {
	name := filepath.Join(director, "newsserver.log")
	f, _, err := zap.Open(name)
	if err != nil {
		return zapcore.AddSync(os.Stderr)
	}
	return zapcore.AddSync(f)
}

func Debugf(format string, args ...interface{}) { sugar.Debugf(format, args...) }
func Infof(format string, args ...interface{})  { sugar.Infof(format, args...) }
func Warnf(format string, args ...interface{})  { sugar.Warnf(format, args...) }
func Errorf(format string, args ...interface{}) { sugar.Errorf(format, args...) }
func Infow(msg string, keysAndValues ...interface{})  { sugar.Infow(msg, keysAndValues...) }
func Errorw(msg string, keysAndValues ...interface{}) { sugar.Errorw(msg, keysAndValues...) }
