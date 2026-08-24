// Package logger provides a structured logger backed by zap.
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with a simpler API.
type Logger struct {
	z *zap.Logger
}

// Field is an alias to zap.Field for convenience.
type Field = zap.Field

// Convenience constructors that mirror zap's API.
var (
	String  = zap.String
	Int     = zap.Int
	Int64   = zap.Int64
	Bool    = zap.Bool
	Any     = zap.Any
	Err     = zap.Error
	Stringer = zap.Stringer
)

// Wrap builds a Logger backed by a caller-supplied *zap.Logger, bypassing
// New's production config. Intended for tests that need to observe log
// output (e.g. via zaptest/observer) without touching stdout/stderr.
func Wrap(z *zap.Logger) *Logger {
	return &Logger{z: z}
}

// New builds a production-ready Logger with the given level string
// (debug, info, warn, error). Defaults to info on invalid input.
func New(level string) *Logger {
	lvl := zapcore.InfoLevel
	if err := lvl.Set(level); err != nil {
		lvl = zapcore.InfoLevel
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	z, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	return &Logger{z: z}
}

// With returns a child Logger with the given fields pre-applied.
func (l *Logger) With(fields ...Field) *Logger {
	return &Logger{z: l.z.With(fields...)}
}

// Info logs a message at info level.
func (l *Logger) Info(msg string, fields ...Field) { l.z.Info(msg, fields...) }

// Warn logs a message at warn level.
func (l *Logger) Warn(msg string, fields ...Field) { l.z.Warn(msg, fields...) }

// Error logs a message at error level.
func (l *Logger) Error(msg string, fields ...Field) { l.z.Error(msg, fields...) }

// Debug logs a message at debug level.
func (l *Logger) Debug(msg string, fields ...Field) { l.z.Debug(msg, fields...) }

// Sync flushes any buffered log entries. Should be called before process exit.
func (l *Logger) Sync() { _ = l.z.Sync() }
