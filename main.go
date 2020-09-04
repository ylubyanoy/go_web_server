package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ylubyanoy/go_web_server/internal"
	"go.uber.org/zap"
)

const servicename = "streamsinfo"

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	appLoger := logger.Sugar().Named(servicename)
	appLoger.Info("The application is starting...")

	appLoger.Info("Reading configuration...")
	port := getEnv("PORT", "8000")
	redisAddr := getEnv("REDIS_URL", "redis://user:@localhost:6379/0")
	appLoger.Info("Configuration is ready")

	shutdown := make(chan error, 2)
	bl := internal.BusinessLogic(appLoger.With("module", "bl"), redisAddr, port, shutdown)
	appLoger.Info("Server are ready")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	select {
	case x := <-interrupt:
		appLoger.Infow("Received", "signal", x.String())
	case err := <-shutdown:
		appLoger.Errorw("Received error from functional unit", "err", err)
	}

	appLoger.Info("Stopping the servers...")
	timeout, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	err := bl.Shutdown(timeout)
	if err != nil {
		appLoger.Errorw("Got an error from the business logic server", "err", err)
	}

	appLoger.Info("The application is stopped.")
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
