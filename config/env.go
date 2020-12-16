package config

import (
	"os"
	"strconv"
	"time"
)

func env(key, defaultValue string) (value string) {
	if value = os.Getenv(key); value == "" {
		value = defaultValue
	}
	return
}

func envDuration(key, defaultValue string) time.Duration {
	value, err := time.ParseDuration(env(key, defaultValue))
	if err != nil {
		panic(err.Error())
	}
	return value
}

func envInt(key, defaultValue string) int {
	value, err := strconv.Atoi(env(key, defaultValue))
	if err != nil {
		panic(err.Error())
	}
	return value
}
