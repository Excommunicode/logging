package logging

import (
	"context"
	"os"
)

func Example_logLevelFromEnv() {
	os.Setenv("LOG_LEVEL", "DEBUG")
	defer os.Unsetenv("LOG_LEVEL")

	SetDefaultLogger(New())
	Info("started")
	Debug("detail", "key", "value")
}

func Example_context() {
	ctx := context.Background()
	ctx = Context(ctx, "request_id", "req-123")
	ctx = Context(ctx, "user_id", 42)

	SetDefaultLogger(New())
	WithContext(ctx).Info("request handled")
}

func Example_fromContext() {
	ctx := context.Background()
	ctx = Context(ctx, "trace_id", "abc")

	data := FromContext(ctx)
	// data is map[string]interface{} with "trace_id" -> "abc"
	_ = data
}
