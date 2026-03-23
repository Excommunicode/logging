package logging

import (
	"context"
	"os"
)

func Example_logLevelFromEnv() {
	os.Setenv("LOG_LEVEL", "DEBUG")
	defer os.Unsetenv("LOG_LEVEL")

	ctx := context.Background()
	ctx = ContextWithLogFields(ctx, "req-1", "tenant-a", "worker-7", "context-manager")

	SetDefaultLogger(New())
	Info(ctx, "started")
	Debug(ctx, "detail", "key", "value")
}

func Example_context() {
	ctx := context.Background()
	ctx = ContextWithLogFields(ctx, "req-123", "tenant-42", "main", "http-handler")
	ctx = Context(ctx, "user_id", 42)

	SetDefaultLogger(New())
	Info(ctx, "request handled", "user_id", 42)
}

func Example_fromContext() {
	ctx := context.Background()
	ctx = Context(ctx, "trace_id", "abc")

	data := FromContext(ctx)
	// data is map[string]interface{} with "trace_id" -> "abc"
	_ = data
}
