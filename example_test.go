package logger_test

import (
	"context"
	"os"

	"github/Excommunicode/logger"
)

func Example_logLevelFromEnv() {
	os.Setenv("LOG_LEVEL", "DEBUG")
	defer os.Unsetenv("LOG_LEVEL")

	log := logger.New()
	log.Info("started")
	log.Debug("detail", "key", "value")
	// OUTPUT:
	// INFO started
	// DEBUG detail key=value
}

func Example_context() {
	ctx := context.Background()
	ctx = logger.Context(ctx, "request_id", "req-123")
	ctx = logger.Context(ctx, "user_id", 42)

	log := logger.New()
	log.WithContext(ctx).Info("request handled")
	// OUTPUT:
	// INFO request handled request_id=req-123 user_id=42
}

func Example_fromContext() {
	ctx := context.Background()
	ctx = logger.Context(ctx, "trace_id", "abc")

	data := logger.FromContext(ctx)
	// data is map[string]interface{} with "trace_id" -> "abc"
	_ = data
}
