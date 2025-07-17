package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log *zap.Logger
)

// InitializeLogger initialize the logger.
func InitializeLogger(logLevel *string, logFile *string) {
	level := parseLogLevel(logLevel)
	file := parseLogFile(logFile)

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{file},
		ErrorOutputPaths: []string{"stderr"},
	}
	var err error
	log, err = cfg.Build()
	if err != nil {
		panic("Failed to initialize logger. err=" + err.Error())
	}
}

func parseLogLevel(logLevel *string) zapcore.Level {
	level := zapcore.InfoLevel
	// get level from env LOG_LEVEL
	logLevelFromEnv := os.Getenv("LOG_LEVEL")
	if logLevelFromEnv != "" {
		logLevel = &logLevelFromEnv
	}
	if logLevel != nil {
		var err error
		level, err = zapcore.ParseLevel(*logLevel)
		if err != nil {
			fmt.Println("Failed to parse log level, use default INFO level. error_level=", *logLevel)
			level = zapcore.InfoLevel
		}
	}
	return level
}

func parseLogFile(logFile *string) string {
	file := "stderr"
	// get log file from env LOG_FILE
	logFileFromEnv := os.Getenv("LOG_FILE")
	if logFileFromEnv != "" {
		logFile = &logFileFromEnv
	}
	if logFile != nil {
		if *logFile == "/dev/stdout" || *logFile == "stdout" {
			file = "stdout"
		} else if *logFile == "/dev/stderr" || *logFile == "stderr" {
			file = "stderr"
		} else {
			file = *logFile
		}
	}
	return file
}

// GetLogger
func GetLogger() *zap.Logger {
	if log == nil {
		InitializeLogger(nil, nil)
	}
	if log == nil {
		panic("Logger is not initialized.")
	}
	return log
}
