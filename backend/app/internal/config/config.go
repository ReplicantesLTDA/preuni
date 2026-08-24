// Package config loads the unified monolith configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config aggregates all settings the monolith needs at startup.
type Config struct {
	Port        string
	LogLevel    string
	DatabaseURL string
	RedisURL    string

	// Auth
	JWTSigningKey        string
	JWTAccessExpirySec   int
	JWTRefreshExpiryDays int

	// User (avatar storage -- MinIO or any S3-compatible endpoint the
	// uploading client can reach directly; presigned URLs are generated
	// against this, the backend never proxies upload bytes)
	StorageEndpoint  string
	StorageAccessKey string
	StorageSecretKey string
	StorageUseSSL    bool
	StorageBucket    string

	// Mail
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPass     string
	SMTPUseTLS   bool
	MailFromAddr string
	MailFromName string
}

// Load reads Config from env. Missing required vars panic at startup.
func Load() Config {
	return Config{
		Port:        getEnv("MONOLITH_PORT", getEnv("PORT", "8080")),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		DatabaseURL: require("DATABASE_URL"),
		RedisURL:    getEnv("REDIS_URL", ""),

		JWTSigningKey:        require("JWT_SIGNING_KEY"),
		JWTAccessExpirySec:   envInt("JWT_ACCESS_EXPIRY_SECONDS", 3600),
		JWTRefreshExpiryDays: envInt("JWT_REFRESH_EXPIRY_DAYS", 30),

		StorageEndpoint:  getEnv("STORAGE_ENDPOINT", "localhost:9000"),
		StorageAccessKey: getEnv("STORAGE_ACCESS_KEY", ""),
		StorageSecretKey: getEnv("STORAGE_SECRET_KEY", ""),
		StorageUseSSL:    getEnv("STORAGE_USE_SSL", "false") == "true",
		StorageBucket:    getEnv("STORAGE_BUCKET", "preuni-avatars"),

		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     envInt("SMTP_PORT", 587),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPass:     getEnv("SMTP_PASS", ""),
		SMTPUseTLS:   envInt("SMTP_PORT", 587) == 465,
		MailFromAddr: getEnv("FROM_EMAIL", "noreply@preuni.com.br"),
		MailFromName: getEnv("MAIL_FROM_NAME", "PreUni"),
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func require(k string) string {
	v := os.Getenv(k)
	if strings.TrimSpace(v) == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", k))
	}
	return v
}
