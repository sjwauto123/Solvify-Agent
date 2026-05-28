package logger

import (
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config 描述 zap 日志和文件轮转配置
type Config struct {
	Level      string
	Filename   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

var (
	mu         sync.Mutex
	log        *zap.Logger
	sugar      *zap.SugaredLogger
	fileWriter *lumberjack.Logger
)

// New 创建 zap 日志实例并初始化全局日志对象
func New(cfg Config) (*zap.Logger, error) {
	mu.Lock()
	defer mu.Unlock()

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

	writers := []zapcore.WriteSyncer{zapcore.AddSync(os.Stdout)}
	if cfg.Filename != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.Filename), 0o755); err != nil {
			return nil, err
		}
		fileWriter = &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    positiveOrDefault(cfg.MaxSize, 100),
			MaxBackups: positiveOrDefault(cfg.MaxBackups, 7),
			MaxAge:     positiveOrDefault(cfg.MaxAge, 30),
			Compress:   cfg.Compress,
		}
		writers = append(writers, zapcore.AddSync(fileWriter))
	} else {
		fileWriter = nil
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(writers...),
		parseLevel(cfg.Level),
	)

	log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))
	sugar = log.Sugar()
	return log, nil
}

// Debug 打印调试日志
func Debug(msg string, fields ...zap.Field) {
	ensureLogger().Debug(msg, fields...)
}

// Info 打印信息日志
func Info(msg string, fields ...zap.Field) {
	ensureLogger().Info(msg, fields...)
}

// Warn 打印警告日志
func Warn(msg string, fields ...zap.Field) {
	ensureLogger().Warn(msg, fields...)
}

// Error 打印错误日志
func Error(msg string, fields ...zap.Field) {
	ensureLogger().Error(msg, fields...)
}

// Fatal 打印致命错误日志并退出进程
func Fatal(msg string, fields ...zap.Field) {
	ensureLogger().Fatal(msg, fields...)
}

// Debugf 打印格式化调试日志
func Debugf(format string, args ...any) {
	ensureSugar().Debugf(format, args...)
}

// Infof 打印格式化信息日志
func Infof(format string, args ...any) {
	ensureSugar().Infof(format, args...)
}

// Warnf 打印格式化警告日志
func Warnf(format string, args ...any) {
	ensureSugar().Warnf(format, args...)
}

// Errorf 打印格式化错误日志
func Errorf(format string, args ...any) {
	ensureSugar().Errorf(format, args...)
}

// Fatalf 打印格式化致命错误日志并退出进程
func Fatalf(format string, args ...any) {
	ensureSugar().Fatalf(format, args...)
}

// With 创建带固定字段的 zap logger
func With(fields ...zap.Field) *zap.Logger {
	return ensureLogger().With(fields...)
}

// Sync 同步日志缓冲区
func Sync() error {
	mu.Lock()
	defer mu.Unlock()

	var syncErr error
	if log != nil {
		syncErr = log.Sync()
	}
	if fileWriter != nil {
		if err := fileWriter.Close(); err != nil && syncErr == nil {
			syncErr = err
		}
		fileWriter = nil
	}
	return syncErr
}

// GetLogger 获取原始 zap logger
func GetLogger() *zap.Logger {
	return ensureLogger()
}

// GetSugaredLogger 获取 sugared logger
func GetSugaredLogger() *zap.SugaredLogger {
	return ensureSugar()
}

// parseLevel 将配置中的日志级别转换为 zap 级别
func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// ensureLogger 返回已初始化 logger 或兜底 logger
func ensureLogger() *zap.Logger {
	mu.Lock()
	defer mu.Unlock()

	if log == nil {
		log = zap.NewNop()
	}
	return log
}

// ensureSugar 返回已初始化 sugared logger 或兜底 sugared logger
func ensureSugar() *zap.SugaredLogger {
	mu.Lock()
	defer mu.Unlock()

	if sugar == nil {
		if log == nil {
			log = zap.NewNop()
		}
		sugar = log.Sugar()
	}
	return sugar
}

// positiveOrDefault 返回正整数配置或默认值
func positiveOrDefault(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}
