package logutil

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogLevel 定义日志级别
type LogLevel string

const (
	// DebugLevel 开发环境日志级别
	DebugLevel LogLevel = "debug"
	// InfoLevel 测试环境日志级别
	InfoLevel LogLevel = "info"
	// WarnLevel 生产环境日志级别
	WarnLevel LogLevel = "warn"
	// ErrorLevel 生产环境日志级别
	ErrorLevel LogLevel = "error"
)

// Logger 日志工具类
type Logger struct {
	*zap.Logger
}

// 全局 Logger 实例
var _defaultLogger *Logger

// InitGlobalLogger 初始化全局 Logger
func InitGlobalLogger(env LogLevel) {
	_defaultLogger = NewLogger(env)
}

// NewLogger 创建日志工具类实例
func NewLogger(env LogLevel) *Logger {
	var config zap.Config

	switch env {
	case DebugLevel:
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case InfoLevel:
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case WarnLevel:
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case ErrorLevel:
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	// 设置日志输出到控制台
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	// 构建 Logger
	logger, err := config.Build()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	return &Logger{logger}
}

// Debug 打印 Debug 级别日志
func Debug(msg string, fields ...zap.Field) {
	_defaultLogger.Logger.Debug(msg, fields...)
}

// Info 打印 Info 级别日志
func Info(msg string, fields ...zap.Field) {
	_defaultLogger.Logger.Info(msg, fields...)
}

// Warn 打印 Warn 级别日志
func Warn(msg string, fields ...zap.Field) {
	_defaultLogger.Logger.Warn(msg, fields...)
}

// Error 打印 Error 级别日志
func Error(msg string, fields ...zap.Field) {
	_defaultLogger.Logger.Error(msg, fields...)
}

// Fatal 打印 Fatal 级别日志并退出程序
func Fatal(msg string, fields ...zap.Field) {
	_defaultLogger.Logger.Fatal(msg, fields...)
}
