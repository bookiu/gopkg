package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// getLogWriter 配置 lumberjack 作为日志输出目标
func getLogWriter(filename string, maxSize, maxBackups, maxAge int, compress bool) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,   // 日志文件路径
		MaxSize:    maxSize,    // 每个日志文件最大大小（MB）
		MaxBackups: maxBackups, // 保留的旧日志文件数量
		MaxAge:     maxAge,     // 旧日志文件保留天数
		Compress:   false,      // 是否压缩旧日志文件
		LocalTime:  true,
	}
	return zapcore.AddSync(lumberJackLogger)
}

// getJsonEncoder 配置日志编码格式（JSON 或 Console）
func getJsonEncoder() zapcore.Encoder {
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
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	return zapcore.NewJSONEncoder(encoderConfig)
}

// createLogger 初始化 zap 日志记录器，支持按级别分文件
func createLogger(infoFile, errorFile string) *zap.Logger {
	// 配置 Info 级别日志的输出
	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.InfoLevel && lvl < zapcore.ErrorLevel
	})

	// 配置 Error 级别日志的输出
	errorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})

	// 配置 lumberjack 写入器
	infoWriter := getLogWriter(infoFile, 10, 3, 7, true)   // Info 日志：10MB 切割，保留 3 个备份，7 天
	errorWriter := getLogWriter(errorFile, 10, 3, 7, true) // Error 日志：10MB 切割，保留 3 个备份，7 天

	// 创建 zap core，支持多输出目标
	core := zapcore.NewTee(
		zapcore.NewCore(getJsonEncoder(), infoWriter, infoLevel),
		zapcore.NewCore(getJsonEncoder(), errorWriter, errorLevel),
		// 可选：同时输出到控制台（开发环境）
		zapcore.NewCore(getJsonEncoder(), zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
	)

	// 创建 logger，添加调用者信息
	logger := zap.New(core, zap.AddCaller())
	return logger
}
