package log

import (
	"context"
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func StartLog() *zap.Logger {

	logger, err := defaultLogger()
	if err != nil {
		log.Fatal("fail start log", err)
		panic(err)
	}

	defer logger.Sync()

	return logger
}

func defaultLogger() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	return cfg.Build()
}

func getLoggerFromContext(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value("logger").(*zap.Logger)

	if !ok {
		return StartLog()
	}

	return logger
}

func LogInfo(ctx context.Context, message string, fields ...zap.Field) {
	logger := getLoggerFromContext(ctx)
	if logger != nil {
		logger.Info(message, fields...)
	}
}

func LogError(ctx context.Context, message string, err error, fields ...zap.Field) {
	logger := getLoggerFromContext(ctx)
	if logger != nil {
		logger.Error(message, append(fields, zap.Error(err))...)
	}
}

func LogWarn(ctx context.Context, message string, fields ...zap.Field) {
	logger := getLoggerFromContext(ctx)
	if logger != nil {
		logger.Warn(message, fields...)
	}
}

func LogDebug(ctx context.Context, message string, fields ...zap.Field) {
	logger := getLoggerFromContext(ctx)
	if logger != nil {
		logger.Debug(message, fields...)
	}
}
