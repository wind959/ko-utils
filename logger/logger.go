package logutil

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	once        sync.Once
	atomicLevel = zap.NewAtomicLevelAt(zapcore.InfoLevel)
)

type LogConfig struct {
	LogLevel          LogLevel  // 日志打印级别 debug  info  warning  error
	LogFormat         LogFormat // 输出到文件日志格式  console, json （仅控制文件格式！）
	LogPath           string    // 输出日志文件路径
	LogFileName       string    // 输出日志文件名称
	LogFileMaxSize    int       // 【日志分割】单个日志文件最多存储量 单位(mb)
	LogFileMaxBackups int       // 【日志分割】日志备份文件最多数量
	LogMaxAge         int       // 日志保留时间，单位: 天 (day)
	LogCompress       bool      // 是否压缩日志
	LogCallerSkip     int       // 调用栈跳过层数
	LogStdout         bool      // 是否输出到控制台
}

// LogFormat 日志输出格式
type LogFormat string

const (
	Console LogFormat = "console"
	Json    LogFormat = "json"
)

// LogLevel 日志打印级别
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

func defaultConfig() LogConfig {
	return LogConfig{
		LogLevel:          LevelInfo,
		LogFileMaxSize:    100,
		LogFileMaxBackups: 7,
		LogMaxAge:         14,
		LogCompress:       true,
		LogCallerSkip:     1,
		LogStdout:         true,    // 默认输出到控制台
		LogFormat:         Console, // 文件格式默认 console（仅影响文件！）
		LogPath:           "",      // 默认不输出到文件
		LogFileName:       "",      // 默认不输出到文件
	}
}

func InitGlobLogger(cfg ...LogConfig) error {
	conf := defaultConfig()
	if len(cfg) > 0 {
		conf = mergeConfig(conf, cfg[0])
	}

	var err error
	once.Do(func() {
		atomicLevel.SetLevel(parseLevel(string(conf.LogLevel)))

		// 构建两个独立的编码器
		consoleEncoder := buildConsoleEncoder()         // 控制台专用（固定格式）
		fileEncoder := buildFileEncoder(conf.LogFormat) // 文件专用（受LogFormat控制）

		// 构建编码器 - 控制台和文件使用不同的编码器
		var cores []zapcore.Core

		// 控制台核心（默认总是开启）
		cores = append(cores, zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			atomicLevel,
		))

		// 文件编码器（根据 LogFormat 决定格式）
		if conf.LogPath != "" && conf.LogFileName != "" {
			filename := filepath.Join(conf.LogPath, conf.LogFileName)

			l := &lumberjack.Logger{
				Filename:   filename,
				MaxSize:    conf.LogFileMaxSize,
				MaxBackups: conf.LogFileMaxBackups,
				MaxAge:     conf.LogMaxAge,
				Compress:   conf.LogCompress,
			}

			cores = append(cores, zapcore.NewCore(
				fileEncoder,
				zapcore.AddSync(l),
				atomicLevel,
			))
		}

		// 合并所有核心
		core := zapcore.NewTee(cores...)

		skip := conf.LogCallerSkip
		if skip <= 0 {
			skip = 1
		}

		logger := zap.New(
			core,
			zap.AddCaller(),
			zap.AddCallerSkip(skip),
			zap.AddStacktrace(zapcore.ErrorLevel),
		)
		zap.ReplaceGlobals(logger)
	})
	return err
}

// SetLevel 设置日志打印级别
func SetLevel(level string) {
	atomicLevel.SetLevel(parseLevel(level))
}

// mergeConfig 合并默认配置和用户配置
func mergeConfig(def, user LogConfig) LogConfig {
	if user.LogLevel != "" {
		def.LogLevel = user.LogLevel
	}
	if user.LogFormat != "" {
		def.LogFormat = user.LogFormat
	}

	// stdout：用户可显式关闭，否则默认开启
	def.LogStdout = user.LogStdout

	// 文件日志：用户必须同时配置路径和文件名才启用
	if user.LogPath != "" && user.LogFileName != "" {
		def.LogPath = user.LogPath
		def.LogFileName = user.LogFileName
	} else {
		def.LogPath = ""
		def.LogFileName = ""
	}

	if user.LogFileMaxSize > 0 {
		def.LogFileMaxSize = user.LogFileMaxSize
	}
	if user.LogFileMaxBackups > 0 {
		def.LogFileMaxBackups = user.LogFileMaxBackups
	}
	if user.LogMaxAge > 0 {
		def.LogMaxAge = user.LogMaxAge
	}

	def.LogCompress = user.LogCompress

	return def
}

func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return zapcore.DebugLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// buildConsoleEncoder 构建控制台编码器（固定格式，带颜色）
func buildConsoleEncoder() zapcore.Encoder {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncodeCaller = zapcore.ShortCallerEncoder
	cfg.EncodeDuration = zapcore.StringDurationEncoder
	// 控制台始终使用大写+颜色
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return zapcore.NewConsoleEncoder(cfg)
}

// buildFileEncoder 构建文件编码器（根据 LogFormat 决定格式）
func buildFileEncoder(format LogFormat) zapcore.Encoder {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncodeCaller = zapcore.ShortCallerEncoder
	cfg.EncodeDuration = zapcore.StringDurationEncoder

	if strings.ToLower(string(format)) == string(Json) {
		// JSON：大写、无颜色
		cfg.EncodeLevel = zapcore.CapitalLevelEncoder
		return zapcore.NewJSONEncoder(cfg)
	}
	// Console：大写但不带颜色（适合文件）
	cfg.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(cfg)
}

func Logger() *zap.Logger {
	return zap.L()
}

func Sugar() *zap.SugaredLogger {
	return zap.S()
}

// Sync 程序退出时，同步日志
func Sync() error {
	return zap.L().Sync()
}

func Replace(logger *zap.Logger) {
	if logger != nil {
		zap.ReplaceGlobals(logger)
	}
}

// 结构化（zap 原生）
func Debug(msg string, fields ...zap.Field) { zap.L().Debug(msg, fields...) }
func Info(msg string, fields ...zap.Field)  { zap.L().Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { zap.L().Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field) { zap.L().Error(msg, fields...) }
func Fatal(msg string, fields ...zap.Field) { zap.L().Fatal(msg, fields...) }
func Panic(msg string, fields ...zap.Field) { zap.L().Panic(msg, fields...) }

// 格式化(Sugar)
func Debugf(format string, args ...any) { zap.S().Debugf(format, args...) }
func Infof(format string, args ...any)  { zap.S().Infof(format, args...) }
func Warnf(format string, args ...any)  { zap.S().Warnf(format, args...) }
func Errorf(format string, args ...any) { zap.S().Errorf(format, args...) }
func Fatalf(format string, args ...any) { zap.S().Fatalf(format, args...) }
func Panicf(format string, args ...any) { zap.S().Panicf(format, args...) }

// 键值对(Sugar)
func Debugw(msg string, keysAndValues ...any) { zap.S().Debugw(msg, keysAndValues...) }
func Infow(msg string, keysAndValues ...any)  { zap.S().Infow(msg, keysAndValues...) }
func Warnw(msg string, keysAndValues ...any)  { zap.S().Warnw(msg, keysAndValues...) }
func Errorw(msg string, keysAndValues ...any) { zap.S().Errorw(msg, keysAndValues...) }
func Fatalw(msg string, keysAndValues ...any) { zap.S().Fatalw(msg, keysAndValues...) }
func Panicw(msg string, keysAndValues ...any) { zap.S().Panicw(msg, keysAndValues...) }
