package main

import (
	"os"
	"testing"

	rollbar "github.com/rollbar/rollbar-go"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	rollbar.Close()
	goleak.VerifyTestMain(m)
}

func TestSetupEnv(t *testing.T) {
	setupEnv()
	port := os.Getenv("PORT")
	assert.NotEqual(t, port, "")
}

func TestSetupRollbar(t *testing.T) {
	setupRollbar()
	assert.Equal(t, rollbar.Token(), os.Getenv("ROLLBAR_SERVER_TOKEN"))
	assert.Equal(t, rollbar.Environment(), os.Getenv("ENVIRONMENT"))
}

func TestGetLogger(t *testing.T) {
	t.Setenv("ENVIRONMENT", "development")
	logger := getLogger()
	assert.NotNil(t, logger)

	t.Setenv("ENVIRONMENT", "production")
	logger = getLogger()
	assert.NotNil(t, logger)
}
