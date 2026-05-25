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
	Port                 string
	LogLevel             string
	DatabaseURL          string
	RedisURL             string
	InternalServiceToken string

	// Auth
	JWTSigningKey        string
	JWTAccessExpirySec   int
	JWTRefreshExpiryDays int

	// User
	S3Bucket string
	S3Region string

	// Self-URL: where in-process services find the monolith's own internal endpoints.
	// In monolith mode, auth handlers self-HTTP to mail+user via this URL.
	SelfBaseURL string

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
	port := getEnv("MONOLITH_PORT", getEnv("PORT", "8080"))
	selfDefault := "http://localhost:" + port
	return Config{
		Port:                 port,
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		DatabaseURL:          require("DATABASE_URL"),
		RedisURL:             getEnv("REDIS_URL", ""),
		InternalServiceToken: require("INTERNAL_SERVICE_TOKEN"),

		JWTSigningKey:        require("JWT_SIGNING_KEY"),
		JWTAccessExpirySec:   envInt("JWT_ACCESS_EXPIRY_SECONDS", 3600),
		JWTRefreshExpiryDays: envInt("JWT_REFRESH_EXPIRY_DAYS", 30),

		S3Bucket: getEnv("S3_BUCKET", "preuni-avatars"),
		S3Region: getEnv("S3_REGION", "us-east-1"),

		SelfBaseURL: getEnv("SELF_BASE_URL", selfDefault),

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
