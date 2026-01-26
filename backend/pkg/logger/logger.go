package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 全局日志实例
var log *zap.SugaredLogger

// Config 日志配置
type Config struct {
	Level  string // 日志级别: debug, info, warn, error
	Format string // 输出格式: json, console
	Output string // 输出路径: stdout 或文件路径
}

// Init 初始化日志系统
func Init(cfg *Config) error {
	// 解析日志级别
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return err
	}

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建编码器
	var encoder zapcore.Encoder
	if cfg.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 创建输出
	var writeSyncer zapcore.WriteSyncer
	if cfg.Output == "" || cfg.Output == "stdout" {
		writeSyncer = zapcore.AddSync(os.Stdout)
	} else {
		// 确保日志目录存在
		logDir := filepath.Dir(cfg.Output)
		if err = os.MkdirAll(logDir, 0755); err != nil {
			return err
		}

		// 打开日志文件
		file, err := os.OpenFile(cfg.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		// 同时输出到文件和控制台
		writeSyncer = zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(file),
			zapcore.AddSync(os.Stdout),
		)
	}

	// 创建核心
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建日志实例
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	log = logger.Sugar()

	return nil
}

// parseLevel 解析日志级别字符串
func parseLevel(levelStr string) (zapcore.Level, error) {
	var level zapcore.Level
	err := level.UnmarshalText([]byte(levelStr))
	return level, err
}

// Debug 输出调试日志
func Debug(args ...interface{}) {
	log.Debug(args...)
}

// Debugf 输出格式化调试日志
func Debugf(template string, args ...interface{}) {
	log.Debugf(template, args...)
}

// Info 输出信息日志
func Info(args ...interface{}) {
	log.Info(args...)
}

// Infof 输出格式化信息日志
func Infof(template string, args ...interface{}) {
	log.Infof(template, args...)
}

// Warn 输出警告日志
func Warn(args ...interface{}) {
	log.Warn(args...)
}

// Warnf 输出格式化警告日志
func Warnf(template string, args ...interface{}) {
	log.Warnf(template, args...)
}

// Error 输出错误日志
func Error(args ...interface{}) {
	log.Error(args...)
}

// Errorf 输出格式化错误日志
func Errorf(template string, args ...interface{}) {
	log.Errorf(template, args...)
}

// Fatal 输出致命错误日志并退出
func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

// Fatalf 输出格式化致命错误日志并退出
func Fatalf(template string, args ...interface{}) {
	log.Fatalf(template, args...)
}

// WithFields 创建带有字段的日志
func WithFields(fields map[string]interface{}) *zap.SugaredLogger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return log.With(args...)
}

// Sync 同步日志缓冲
func Sync() error {
	if log != nil {
		return log.Sync()
	}
	return nil
}
