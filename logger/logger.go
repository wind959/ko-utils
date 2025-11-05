package logutil

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"
	"time"
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

// 全局 Logger 实例
var _defaultLogger *Logger

// Logger 日志工具类
type Logger struct {
	*zap.Logger
	sugar *zap.SugaredLogger
}

// InitGlobalLogger 初始化全局 Logger
func InitGlobalLogger(env LogLevel) {
	_defaultLogger = NewLogger(env)
}

// NewLogger 创建日志工具类实例
func NewLogger(env LogLevel) *Logger {
	level := strings.ToLower(string(env))
	var config zap.Config

	switch level {
	case "debug":
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}
	// 自定义 EncoderConfig
	config.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:       "time",                           // 日志中时间字段的键名
		LevelKey:      "level",                          // 日志中级别字段的键名
		NameKey:       "logger",                         // 日志中记录器名称字段的键名
		CallerKey:     "caller",                         // 日志中调用者信息字段的键名
		FunctionKey:   zapcore.OmitKey,                  // 日志中函数名字段的键名，这里设置为忽略
		MessageKey:    "msg",                            // 日志中消息字段的键名
		StacktraceKey: "stacktrace",                     // 日志中堆栈跟踪字段的键名
		LineEnding:    zapcore.DefaultLineEnding,        // 日志行的结束符，默认为 "\n"
		EncodeLevel:   zapcore.CapitalColorLevelEncoder, // 日志级别的编码方式，这里使用带颜色的编码
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		EncodeDuration: zapcore.StringDurationEncoder, // 日志中持续时间的编码方式，这里使用字符串格式
		EncodeCaller:   zapcore.ShortCallerEncoder,    // 日志中调用者信息的编码方式，这里使用短格式
	}

	// 设置日志输出到控制台
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	// 构建 Logger
	logger, err := config.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	return &Logger{
		Logger: logger,
		sugar:  logger.Sugar(),
	}
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

func Panic(msg string, fields ...zap.Field) {
	_defaultLogger.Logger.Panic(msg, fields...)
}

// Fatal 打印 Fatal 级别日志并退出程序
func Fatal(msg string, fields ...zap.Field) {
	_defaultLogger.Logger.Fatal(msg, fields...)
}

/****************** 格式化日志方法 ******************/
func Debugf(format string, args ...interface{}) {
	_defaultLogger.sugar.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	_defaultLogger.sugar.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	_defaultLogger.sugar.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	_defaultLogger.sugar.Errorf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	_defaultLogger.sugar.Panicf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	_defaultLogger.sugar.Fatalf(format, args...)
}
