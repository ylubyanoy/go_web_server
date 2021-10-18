package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"

	"github.com/ylubyanoy/go_web_server/internal"
	"github.com/ylubyanoy/go_web_server/internal/config"
	"github.com/ylubyanoy/go_web_server/internal/data/postgres"
	"github.com/ylubyanoy/go_web_server/internal/storages/redis_store"
	"github.com/ylubyanoy/go_web_server/internal/utils"
)

const servicename = "streamsinfo"

func main() {

	appLoger := utils.NewLogger(servicename)
	appLoger.Info("The application is starting...")

	appLoger.Info("Reading configuration...")

	cfg := config.NewConfig()
	err := cleanenv.ReadConfig("configs/config.yml", cfg)
	if err != nil {
		appLoger.Fatalw("Can't read config", zap.Error(err))
	}

	appLoger.Info("Configuration is ready")

	repo, err := postgres.NewPostgresRepository(cfg.RepoURL, appLoger.With("module", "pgstore"))
	if err != nil {
		appLoger.Fatalw("Can't connect to pgstore", zap.Error(err))
	}
	appLoger.Info("Connected to pgstore")

	sessManager, err := redis_store.New(cfg.RedisURL)
	if err != nil {
		appLoger.Fatalw("Can't connect to storage", zap.Error(err))
	}
	appLoger.Info("Connected to Redis")

	shutdown := make(chan error, 2)
	bl := internal.BusinessLogic(appLoger.With("module", "bl"), sessManager, cfg.Port, repo, shutdown)
	appLoger.Info("Server are ready")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	select {
	case x := <-interrupt:
		appLoger.Infow("Received", "signal", x.String())
	case err := <-shutdown:
		appLoger.Errorw("Received error from functional unit", zap.Error(err))
	}

	appLoger.Info("Stopping the servers...")
	timeout, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	err = bl.Shutdown(timeout)
	if err != nil {
		appLoger.Errorw("Got an error from the business logic server", zap.Error(err))
	}

	appLoger.Info("The application is stopped.")
}
