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

// var streamers = `{
// 	"users": [
// 	  {"username": "thaina_"},
// 	  {"username": "blabalbee"},
// 	  {"username": "Smorodinova"},
// 	  {"username": "CekLena"},
// 	  {"username": "JowyBear"},
// 	  {"username": "pimpka74"},
// 	  {"username": "icytoxictv"},
// 	  {"username": "ustepuka"},
// 	  {"username": "AlenochkaBT"},
// 	  {"username": "ViktoriiShka"},
// 	  {"username": "irenchik"},
// 	  {"username": "lola_grrr"},
// 	  {"username": "Sensoria"},
// 	  {"username": "aisumaisu"},
// 	  {"username": "PANGCHOM"},
// 	  {"username": "Danucd"}
// 	]
// }`

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	appLoger := logger.Sugar().Named(servicename)
	appLoger.Info("The application is starting...")

	// redisAddr := getEnv("REDIS_URL", "redis://user:@localhost:6379/0")
	// redisConn := &redis.Pool{
	// 	MaxIdle:     10,
	// 	IdleTimeout: 240 * time.Second,
	// 	Dial: func() (redis.Conn, error) {
	// 		redisConn, err := redis.DialURL(redisAddr)
	// 		if err != nil {
	// 			appLoger.Fatalw("Can't connect to Redis", "err", err)
	// 		}
	// 		return redisConn, nil
	// 	},
	// }
	// sessManager = NewConnManager(redisConn)
	// rc := redisConn.Get()
	// _, err := redis.String(rc.Do("PING"))
	// if err != nil {
	// 	appLoger.Fatalw("Can't connect to Redis", "err", err)
	// }
	// rc.Close()
	// appLoger.Info("Connected to Redis")

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
