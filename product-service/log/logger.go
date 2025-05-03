package logger

import "go.uber.org/zap"

var Logger *zap.SugaredLogger
var baseLogger *zap.Logger

func InitLogger() {
	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	baseLogger = l
	Logger = l.Sugar()
}

func Info(msg string, fields ...zap.Field) {
	baseLogger.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	baseLogger.Error(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	baseLogger.Debug(msg, fields...)
}

func Err(msg string, err error, fields ...zap.Field) {
	baseLogger.Error(msg, append(fields, zap.Error(err))...)
}

func InfoE(msg string, err error, fields ...zap.Field) {
	baseLogger.Info(msg, append(fields, zap.Error(err))...)
}

func DebugE(msg string, err error, fields ...zap.Field) {
	baseLogger.Debug(msg, append(fields, zap.Error(err))...)
}

func Sync() {
	if baseLogger != nil {
		_ = baseLogger.Sync()
	}
	if Logger != nil {
		_ = Logger.Sync()
	}
}

func Str(key, val string) zap.Field {
	return zap.String(key, val)
}

func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

func ErrField(err error) zap.Field {
	return zap.Error(err)
}
