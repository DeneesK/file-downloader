package logger

import (
	"go.uber.org/zap"
)

const (
	dev  = "dev"
	prod = "prod"
)

func NewLogger(env string) *zap.SugaredLogger {
	var logger *zap.Logger
	var err error

	switch env {
	case dev:
		logger, err = zap.NewDevelopment()
	case prod:
		logger, err = zap.NewProduction()
	}

	if err != nil {
		logger.Fatal("failed to initialized new logger", zap.String("err", err.Error()))
	}

	sugar := logger.Sugar()
	return sugar
}
