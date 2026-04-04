// Package config loads service configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strings"
)

// Base holds configuration fields common across all services.
type Base struct {
	Port                  string
	LogLevel              string
	DatabaseURL           string
	RedisURL              string
	InternalServiceToken  string
}

// Load reads the Base configuration from environment variables.
// Required variables that are missing cause a panic with a descriptive message
// so that misconfigured deployments fail loudly at startup.
func Load() Base {
	return Base{
		Port:                 getEnv("PORT", "8080"),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		DatabaseURL:          require("DATABASE_URL"),
		RedisURL:             getEnv("REDIS_URL", ""),
		InternalServiceToken: require("INTERNAL_SERVICE_TOKEN"),
	}
}

// getEnv returns the value of the named environment variable or the default.
func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

// require returns the value of the named environment variable or panics.
func require(key string) string {
	v := os.Getenv(key)
	if strings.TrimSpace(v) == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}
