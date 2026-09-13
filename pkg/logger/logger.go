package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(debug bool) *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.Sampling = nil
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logFile := getLogFilePath()
	cfg.OutputPaths = []string{"stderr", logFile}
	cfg.ErrorOutputPaths = []string{"stderr", logFile}

	cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	if debug {
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}

	l, err := cfg.Build(zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		panic(fmt.Sprintf("failed to build logger: %v", err))
	}
	return l
}

func getLogFilePath() string {
	const logDir = "logs"
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		panic(fmt.Sprintf("failed to create log directory: %v", err))
	}
	return filepath.Join(logDir, "prepper.log")
}
