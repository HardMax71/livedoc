package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

func init() {
	var err error
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	log, err = config.Build()
	if err != nil {
		panic(err)
	}
}

func Logger() *zap.Logger {
	return log
}

func Sugar() *zap.SugaredLogger {
	return log.Sugar()
}
