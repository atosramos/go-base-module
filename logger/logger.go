// Package logger provides a structured, tenant-aware logging facility
// for the auth-service, built on top of uber-go/zap.
//
// Ref: gemini/agents/auth_agent.md — §1.A (Zero-Trust Multi-tenant)
// Ref: gemini/rules/restag_auth_rules.md — §6 (Logs de Auditoria do RBAC)
//
// SECURITY: Never log sensitive authentication details (passwords, JWT private keys,
// TOTP secrets, session tokens) at any logging level.
package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps a zap.Logger and provides convenience methods for
// tenant-scoped and auth-domain-specific structured logging.
type Logger struct {
	zl *zap.Logger
}

// New creates a new Logger. In production mode, outputs JSON. In development
// mode, outputs human-readable colored console format.
func New(production bool) (*Logger, error) {
	var cfg zap.Config

	if production {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "ts"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.EncoderConfig.LevelKey = "level"
		cfg.EncoderConfig.MessageKey = "msg"
		cfg.EncoderConfig.CallerKey = "caller"
		cfg.EncoderConfig.StacktraceKey = "stacktrace"

		levelStr := os.Getenv("LOG_LEVEL")
		if levelStr != "" {
			level, err := zapcore.ParseLevel(levelStr)
			if err == nil {
				cfg.Level = zap.NewAtomicLevelAt(level)
			}
		}
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	zl, err := cfg.Build(
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}

	return &Logger{zl: zl}, nil
}

// Sync flushes any buffered log entries. Must be called via defer in main().
func (l *Logger) Sync() {
	// Error intentionally ignored — common on stdout/stderr
	_ = l.zl.Sync()
}

// WithTenant returns a child Logger pre-populated with tenant_id field.
// MUST be called to ensure all subsequent log entries are tenant-scoped.
//
// Ref: gemini/agents/auth_agent.md — §1.A (Zero-Trust Multi-tenant)
func (l *Logger) WithTenant(tenantID string) *Logger {
	return &Logger{
		zl: l.zl.With(
			zap.String("tenant_id", tenantID),
		),
	}
}

// WithRequestID returns a child Logger with the HTTP request ID for distributed tracing.
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		zl: l.zl.With(zap.String("request_id", requestID)),
	}
}

// WithError returns a child Logger pre-populated with an error field.
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		zl: l.zl.With(zap.Error(err)),
	}
}

// Debug logs a message at debug level. Not emitted in production.
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zl.Debug(msg, fields...)
}

// Info logs an informational message.
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zl.Info(msg, fields...)
}

// Warn logs a warning. Used for security exceptions that do not fail the service but warrant notice.
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zl.Warn(msg, fields...)
}

// Error logs an error. Used for critical errors or failed transactions.
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zl.Error(msg, fields...)
}

// Fatal logs an error and terminates the process. Used only during startup failures.
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.zl.Fatal(msg, fields...)
}

// Named returns a child logger with a subsystem name appended to the logger name.
// Example: logger.Named("token_manager"), logger.Named("totp_authenticator")
func (l *Logger) Named(name string) *Logger {
	return &Logger{zl: l.zl.Named(name)}
}

// Zap returns the underlying zap.Logger for use with third-party libraries
// that accept *zap.Logger directly.
func (l *Logger) Zap() *zap.Logger {
	return l.zl
}

// String is a convenience alias for zap.String.
func String(key, val string) zap.Field { return zap.String(key, val) }

// Any is a convenience alias for zap.Any.
func Any(key string, val interface{}) zap.Field { return zap.Any(key, val) }

// Int is a convenience alias for zap.Int.
func Int(key string, val int) zap.Field { return zap.Int(key, val) }

// Bool is a convenience alias for zap.Bool.
func Bool(key string, val bool) zap.Field { return zap.Bool(key, val) }
