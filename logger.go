package logging

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// Level represents log level.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
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
		return "INFO"
	}
}

// ParseLevel parses string to Level. Unknown values default to LevelInfo.
func ParseLevel(s string) Level {
	switch s {
	case "DEBUG", "debug":
		return LevelDebug
	case "INFO", "info", "":
		return LevelInfo
	case "WARN", "warn":
		return LevelWarn
	case "ERROR", "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// ILogger is the interface for leveled logging. *Logger and wrapped loggers implement it.
type ILogger interface {
	Debug(msg string, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Warn(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})
	WithContext(ctx context.Context) ILogger
	With(keyvals ...interface{}) ILogger
}

// Logger writes leveled log messages.
type Logger struct {
	level Level
	mu    sync.Mutex
}

// Ensure *Logger implements ILogger.
var _ ILogger = (*Logger)(nil)

var (
	defaultLoggerMu sync.RWMutex
	defaultLogger   = New()
)

// levelFromEnv reads LOG_LEVEL from environment and returns Level.
func levelFromEnv() Level {
	return ParseLevel(os.Getenv("LOG_LEVEL"))
}

// New returns a logger with level from LOG_LEVEL env.
func New() *Logger {
	return &Logger{level: levelFromEnv()}
}

// NewWithLevel returns a logger with the given level (ignores env).
func NewWithLevel(level Level) *Logger {
	return &Logger{level: level}
}

// SetDefaultLogger replaces package-level logger used by functional API.
func SetDefaultLogger(l *Logger) {
	if l == nil {
		return
	}
	defaultLoggerMu.Lock()
	defer defaultLoggerMu.Unlock()
	defaultLogger = l
}

func getDefaultLogger() *Logger {
	defaultLoggerMu.RLock()
	defer defaultLoggerMu.RUnlock()
	return defaultLogger
}

// SetLevel sets level for package-level logger.
func SetLevel(level Level) {
	getDefaultLogger().SetLevel(level)
}

// CurrentLevel returns level for package-level logger.
func CurrentLevel() Level {
	return getDefaultLogger().Level()
}

// SetLevel sets the minimum log level at runtime.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Level returns current level.
func (l *Logger) Level() Level {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

func (l *Logger) enabled(level Level) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return level >= l.level
}

func (l *Logger) log(level Level, msg string, keyvals ...interface{}) {
	if !l.enabled(level) {
		return
	}

	out := formatLogLine(level, msg, keyvals...)
	fmt.Println(out)
}

func formatLogLine(level Level, msg string, keyvals ...interface{}) string {
	ts := time.Now().Format("2006-01-02T15:04:05.000")
	fields := map[string]interface{}{
		"request_id": "-",
		"tenant_id":  "-",
		"thread":     "-",
		"class":      "context-manager",
	}

	for i := 0; i+1 < len(keyvals); i += 2 {
		k := fmt.Sprint(keyvals[i])
		fields[k] = keyvals[i+1]
	}

	parts := []string{
		fmt.Sprintf("[%s]", ts),
		fmt.Sprintf("[%s]", level.String()),
		fmt.Sprintf("[request_id=%v]", fields["request_id"]),
		fmt.Sprintf("[tenant_id=%v]", fields["tenant_id"]),
		fmt.Sprintf("[thread=%v]", fields["thread"]),
		fmt.Sprintf("[class=%v]", fields["class"]),
	}

	if msg != "" {
		parts = append(parts, msg)
	}

	for i := 0; i+1 < len(keyvals); i += 2 {
		k := fmt.Sprint(keyvals[i])
		if k == "request_id" || k == "tenant_id" || k == "thread" || k == "class" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%v=%v", keyvals[i], keyvals[i+1]))
	}

	return strings.Join(parts, " ")
}

// Debug logs at LevelDebug.
func (l *Logger) Debug(msg string, keyvals ...interface{}) {
	l.log(LevelDebug, msg, keyvals...)
}

// Info logs at LevelInfo.
func (l *Logger) Info(msg string, keyvals ...interface{}) {
	l.log(LevelInfo, msg, keyvals...)
}

// Warn logs at LevelWarn.
func (l *Logger) Warn(msg string, keyvals ...interface{}) {
	l.log(LevelWarn, msg, keyvals...)
}

// Error logs at LevelError.
func (l *Logger) Error(msg string, keyvals ...interface{}) {
	l.log(LevelError, msg, keyvals...)
}

// Debug logs using fields extracted from ctx and additional keyvals.
func Debug(ctx context.Context, msg string, keyvals ...interface{}) {
	all := append(contextLogFields(ctx), keyvals...)
	getDefaultLogger().Debug(msg, all...)
}

// Info logs using fields extracted from ctx and additional keyvals.
func Info(ctx context.Context, msg string, keyvals ...interface{}) {
	all := append(contextLogFields(ctx), keyvals...)
	getDefaultLogger().Info(msg, all...)
}

// Warn logs using fields extracted from ctx and additional keyvals.
func Warn(ctx context.Context, msg string, keyvals ...interface{}) {
	all := append(contextLogFields(ctx), keyvals...)
	getDefaultLogger().Warn(msg, all...)
}

// Error logs using fields extracted from ctx and additional keyvals.
func Error(ctx context.Context, msg string, keyvals ...interface{}) {
	all := append(contextLogFields(ctx), keyvals...)
	getDefaultLogger().Error(msg, all...)
}

// WithContext returns contextual logger based on package-level default logger.
func WithContext(ctx context.Context) ILogger {
	return getDefaultLogger().WithContext(ctx)
}

// With returns logger with static fields based on package-level default logger.
func With(keyvals ...interface{}) ILogger {
	return getDefaultLogger().With(keyvals...)
}
