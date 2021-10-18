package utils

import "go.uber.org/zap"

func NewLogger(lName string) *zap.SugaredLogger {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	return logger.Sugar().Named(lName)
}
