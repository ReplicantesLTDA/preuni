package config

import "testing"

// TestLoad_UsesDefaultsWhenOptionalVarsAreUnset covers getEnv/envInt's
// default-value branches and Load's overall happy path -- Load had 0%
// coverage before (never unit tested directly).
func TestLoad_UsesDefaultsWhenOptionalVarsAreUnset(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost/db")
	t.Setenv("JWT_SIGNING_KEY", "secret")
	for _, k := range []string{
		"MONOLITH_PORT", "PORT", "LOG_LEVEL", "REDIS_URL",
		"JWT_ACCESS_EXPIRY_SECONDS", "JWT_REFRESH_EXPIRY_DAYS",
		"S3_BUCKET", "S3_REGION", "SMTP_HOST", "SMTP_PORT",
		"SMTP_USER", "SMTP_PASS", "FROM_EMAIL", "MAIL_FROM_NAME",
	} {
		t.Setenv(k, "")
	}

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Port)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want info", cfg.LogLevel)
	}
	if cfg.DatabaseURL != "postgres://u:p@localhost/db" {
		t.Errorf("DatabaseURL = %q", cfg.DatabaseURL)
	}
	if cfg.JWTSigningKey != "secret" {
		t.Errorf("JWTSigningKey = %q", cfg.JWTSigningKey)
	}
	if cfg.JWTAccessExpirySec != 3600 {
		t.Errorf("JWTAccessExpirySec = %d, want 3600", cfg.JWTAccessExpirySec)
	}
	if cfg.JWTRefreshExpiryDays != 30 {
		t.Errorf("JWTRefreshExpiryDays = %d, want 30", cfg.JWTRefreshExpiryDays)
	}
	if cfg.S3Bucket != "preuni-avatars" {
		t.Errorf("S3Bucket = %q", cfg.S3Bucket)
	}
	if cfg.SMTPPort != 587 {
		t.Errorf("SMTPPort = %d, want 587", cfg.SMTPPort)
	}
	if cfg.SMTPUseTLS {
		t.Error("SMTPUseTLS should be false for the default port 587")
	}
	if cfg.MailFromAddr != "noreply@preuni.com.br" {
		t.Errorf("MailFromAddr = %q", cfg.MailFromAddr)
	}
}

// TestLoad_UsesProvidedVarsWhenSet covers getEnv/envInt's "var is set"
// branches, including MONOLITH_PORT taking priority over PORT, and
// SMTPUseTLS's true branch (port 465).
func TestLoad_UsesProvidedVarsWhenSet(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost/db")
	t.Setenv("JWT_SIGNING_KEY", "secret")
	t.Setenv("MONOLITH_PORT", "9090")
	t.Setenv("PORT", "3000")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("JWT_ACCESS_EXPIRY_SECONDS", "120")
	t.Setenv("JWT_REFRESH_EXPIRY_DAYS", "7")
	t.Setenv("S3_BUCKET", "custom-bucket")
	t.Setenv("S3_REGION", "eu-west-1")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "465")
	t.Setenv("SMTP_USER", "user")
	t.Setenv("SMTP_PASS", "pass")
	t.Setenv("FROM_EMAIL", "hi@preuni.com.br")
	t.Setenv("MAIL_FROM_NAME", "Custom Name")

	cfg := Load()

	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want 9090 (MONOLITH_PORT takes priority over PORT)", cfg.Port)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want debug", cfg.LogLevel)
	}
	if cfg.RedisURL != "redis://localhost:6379" {
		t.Errorf("RedisURL = %q", cfg.RedisURL)
	}
	if cfg.JWTAccessExpirySec != 120 {
		t.Errorf("JWTAccessExpirySec = %d, want 120", cfg.JWTAccessExpirySec)
	}
	if cfg.JWTRefreshExpiryDays != 7 {
		t.Errorf("JWTRefreshExpiryDays = %d, want 7", cfg.JWTRefreshExpiryDays)
	}
	if cfg.S3Bucket != "custom-bucket" {
		t.Errorf("S3Bucket = %q", cfg.S3Bucket)
	}
	if cfg.SMTPPort != 465 {
		t.Errorf("SMTPPort = %d, want 465", cfg.SMTPPort)
	}
	if !cfg.SMTPUseTLS {
		t.Error("SMTPUseTLS should be true for port 465")
	}
	if cfg.MailFromName != "Custom Name" {
		t.Errorf("MailFromName = %q", cfg.MailFromName)
	}
}

// TestLoad_PanicsWhenRequiredVarIsMissing covers require()'s panic branch.
func TestLoad_PanicsWhenRequiredVarIsMissing(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_SIGNING_KEY", "secret")

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected Load() to panic when DATABASE_URL is unset")
		}
	}()
	Load()
}

// TestEnvInt_FallsBackOnUnparseableValue covers envInt's strconv.Atoi
// error branch (a non-numeric value falls back to the default).
func TestEnvInt_FallsBackOnUnparseableValue(t *testing.T) {
	t.Setenv("SOME_INT_VAR", "not-a-number")
	if got := envInt("SOME_INT_VAR", 42); got != 42 {
		t.Errorf("envInt with an unparseable value = %d, want fallback 42", got)
	}
}

// TestRequire_PanicsOnWhitespaceOnlyValue covers require()'s
// strings.TrimSpace check (whitespace-only counts as unset).
func TestRequire_PanicsOnWhitespaceOnlyValue(t *testing.T) {
	t.Setenv("WHITESPACE_ONLY_VAR", "   ")
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected require() to panic for a whitespace-only value")
		}
	}()
	require("WHITESPACE_ONLY_VAR")
}
