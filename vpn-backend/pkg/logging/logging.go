package logging

import (
	"context"
	"log"
	"os"
	"sync"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger struct {
	level  LogLevel
	logger *log.Logger
	mu     sync.Mutex
}

var (
	defaultLogger *Logger
	once          sync.Once
)

func Init(level LogLevel) {
	once.Do(func() {
		defaultLogger = &Logger{
			level:  level,
			logger: log.New(os.Stdout, "", log.LstdFlags|log.LUTC),
		}
	})
}

func Get() *Logger {
	if defaultLogger == nil {
		Init(LevelInfo)
	}
	return defaultLogger
}

func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) log(level LogLevel, msg string, fields map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	// Sanitize fields - remove any sensitive data
	safeFields := sanitizeFields(fields)

	l.logger.Printf("[%s] %s %v", level.String(), msg, safeFields)
}

func sanitizeFields(fields map[string]interface{}) map[string]interface{} {
	if fields == nil {
		return nil
	}

	sensitiveKeys := map[string]bool{
		"password":           true,
		"token":              true,
		"access_token":       true,
		"refresh_token":      true,
		"private_key":        true,
		"secret":             true,
		"authorization":      true,
		"cookie":             true,
		"purchase_token":     true,
		"order_id":           true,
		"email":              true, // PII - hash in production
		"user_id":            true, // PII - hash in production
		"ip":                 true, // PII
		"client_ip":          true,
	}

	result := make(map[string]interface{})
	for k, v := range fields {
		if sensitiveKeys[k] {
			result[k] = "[REDACTED]"
		} else {
			result[k] = v
		}
	}
	return result
}

func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	l.log(LevelDebug, msg, fields)
}

func (l *Logger) Info(msg string, fields map[string]interface{}) {
	l.log(LevelInfo, msg, fields)
}

func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	l.log(LevelWarn, msg, fields)
}

func (l *Logger) Error(msg string, fields map[string]interface{}) {
	l.log(LevelError, msg, fields)
}

// Context-aware logging
type ctxKey string

const loggerKey ctxKey = "logger"

func WithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(loggerKey).(*Logger); ok {
		return logger
	}
	return Get()
}

// Structured logging helpers for common operations
func LogAuthAttempt(ctx context.Context, email string, success bool) {
	FromContext(ctx).Info("auth_attempt", map[string]interface{}{
		"success": success,
	})
}

func LogVPNConnect(ctx context.Context, serverID string, success bool) {
	FromContext(ctx).Info("vpn_connect", map[string]interface{}{
		"server_id": serverID,
		"success":   success,
	})
}

func LogConfigRequest(ctx context.Context, serverID string, hasSubscription bool) {
	FromContext(ctx).Info("config_request", map[string]interface{}{
		"server_id":         serverID,
		"has_subscription":  hasSubscription,
	})
}

func LogBillingEvent(ctx context.Context, event string, success bool) {
	FromContext(ctx).Info("billing_event", map[string]interface{}{
		"event":   event,
		"success": success,
	})
}

func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}