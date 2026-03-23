package logging

import (
	"context"
)

type contextKey struct{}

const (
	LogFieldRequestID = "request_id"
	LogFieldTenantID  = "tenant_id"
	LogFieldThread    = "thread"
	LogFieldClass     = "class"
)

// Context stores key in ctx and returns new context. Use for adding logging fields to context.
// Later extract them with FromContext and pass to Logger.WithContext or include in log calls.
func Context(ctx context.Context, key string, value interface{}) context.Context {
	m := fromContext(ctx)
	if m == nil {
		m = make(map[string]interface{})
	}
	// Copy to avoid mutating parent context's map
	m2 := make(map[string]interface{}, len(m)+1)
	for k, v := range m {
		m2[k] = v
	}
	m2[key] = value
	return context.WithValue(ctx, contextKey{}, m2)
}

// FromContext extracts logging data from context. Returns nil if nothing was set.
func FromContext(ctx context.Context) map[string]interface{} {
	return fromContext(ctx)
}

// ContextWithLogFields stores standard logging fields in context.
// Empty values are ignored.
func ContextWithLogFields(ctx context.Context, requestID, tenantID, thread, className string) context.Context {
	if requestID != "" {
		ctx = Context(ctx, LogFieldRequestID, requestID)
	}
	if tenantID != "" {
		ctx = Context(ctx, LogFieldTenantID, tenantID)
	}
	if thread != "" {
		ctx = Context(ctx, LogFieldThread, thread)
	}
	if className != "" {
		ctx = Context(ctx, LogFieldClass, className)
	}
	return ctx
}

func contextLogFields(ctx context.Context) []interface{} {
	m := fromContext(ctx)
	if len(m) == 0 {
		return nil
	}

	fields := make([]interface{}, 0, 8)
	appendIfPresent := func(key string) {
		if v, ok := m[key]; ok {
			fields = append(fields, key, v)
		}
	}

	appendIfPresent(LogFieldRequestID)
	appendIfPresent(LogFieldTenantID)
	appendIfPresent(LogFieldThread)
	appendIfPresent(LogFieldClass)
	return fields
}

func fromContext(ctx context.Context) map[string]interface{} {
	v := ctx.Value(contextKey{})
	if v == nil {
		return nil
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	return m
}

// WithContext returns a logger that includes all key-value pairs from ctx (set via Context) in every log.
func (l *Logger) WithContext(ctx context.Context) ILogger {
	m := FromContext(ctx)
	if len(m) == 0 {
		return l
	}
	keyvals := make([]interface{}, 0, len(m)*2)
	for k, v := range m {
		keyvals = append(keyvals, k, v)
	}
	return l.With(keyvals...)
}

// With returns a new logger that always adds the given key-value pairs to every log.
func (l *Logger) With(keyvals ...interface{}) ILogger {
	return &loggerWith{parent: l, keyvals: keyvals}
}

// Ensure *loggerWith implements ILogger.
var _ ILogger = (*loggerWith)(nil)

// loggerWith is a logger that prepends keyvals to every log.
type loggerWith struct {
	parent  *Logger
	keyvals []interface{}
}

func (l *loggerWith) log(level Level, msg string, keyvals ...interface{}) {
	all := append(l.keyvals, keyvals...)
	l.parent.log(level, msg, all...)
}

func (l *loggerWith) enabled(level Level) bool { return l.parent.enabled(level) }
func (l *loggerWith) Level() Level             { return l.parent.Level() }
func (l *loggerWith) SetLevel(level Level)     { l.parent.SetLevel(level) }

func (l *loggerWith) Debug(msg string, keyvals ...interface{}) { l.log(LevelDebug, msg, keyvals...) }
func (l *loggerWith) Info(msg string, keyvals ...interface{})  { l.log(LevelInfo, msg, keyvals...) }
func (l *loggerWith) Warn(msg string, keyvals ...interface{})  { l.log(LevelWarn, msg, keyvals...) }
func (l *loggerWith) Error(msg string, keyvals ...interface{}) { l.log(LevelError, msg, keyvals...) }

func (l *loggerWith) WithContext(ctx context.Context) ILogger {
	m := FromContext(ctx)
	if len(m) == 0 {
		return l
	}
	extra := make([]interface{}, 0, len(m)*2)
	for k, v := range m {
		extra = append(extra, k, v)
	}
	return l.parent.With(append(l.keyvals, extra...)...)
}

func (l *loggerWith) With(keyvals ...interface{}) ILogger {
	return l.parent.With(append(l.keyvals, keyvals...)...)
}
