package logging

import (
	"context"
	"fmt"
	"os"
	"sync"
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
	// Simple format: LEVEL message key=value key=value
	out := level.String() + " " + msg
	for i := 0; i+1 < len(keyvals); i += 2 {
		out += fmt.Sprintf(" %v=%v", keyvals[i], keyvals[i+1])
	}
	fmt.Println(out)
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
