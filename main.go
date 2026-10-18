package main

import (
	_ "embed"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/albertyw/reaction-pics/server"
	"github.com/joho/godotenv"
	rollbar "github.com/rollbar/rollbar-go"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//go:embed .env
var dotenv []byte

func setupEnv() {
	envMap, err := godotenv.Unmarshal(string(dotenv))
	if err != nil {
		panic(err)
	}
	for key, value := range envMap {
		err := os.Setenv(key, value)
		if err != nil {
			panic(err)
		}
	}
}

func setupRollbar() {
	rollbar.SetToken(os.Getenv("ROLLBAR_SERVER_TOKEN"))
	rollbar.SetEnvironment(os.Getenv("ENVIRONMENT"))
}

func getLogger() *zap.Logger {
	var config zap.Config
	if os.Getenv("ENVIRONMENT") == "development" {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		filename = "."
	}
	logFile := filepath.Join(filepath.Dir(filename), "logs", "app", "app.log")
	config.OutputPaths = []string{"stdout", logFile}
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	return logger
}

func main() {
	setupEnv()
	setupRollbar()
	logger := getLogger()
	server.Run(logger)
	err := logger.Sync()
	if err != nil {
		rollbar.Error(rollbar.ERR, err)
	}
}
