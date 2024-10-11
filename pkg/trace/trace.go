package trace

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
)

var globalTracer *Tracer

type Tracer struct {
	logFunc LogFunc
}

// LogFunc defines the signature for the logging function
type LogFunc func(msg string, fields ...interface{})

type TracerOption func(*Tracer)

func WithLogFunc(logFunc LogFunc) TracerOption {
	return func(t *Tracer) {
		t.logFunc = logFunc
	}
}

// Init initializes the global tracer and returns an error if initialization fails
func Init(endpoint, serviceName string, options ...TracerOption) error {
	t := &Tracer{
		logFunc: func(msg string, fields ...interface{}) {}, // Use a no-op log function by default
	}

	for _, option := range options {
		option(t)
	}

	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LocalAgentHostPort: endpoint,
		},
	}

	_, err := cfg.InitGlobalTracer(serviceName)
	if err != nil {
		t.logFunc("Failed to init tracing", "error", err)
		return fmt.Errorf("failed to init tracing: %w", err)
	}

	t.logFunc("Tracing initialized", "endpoint", endpoint, "service", serviceName)
	globalTracer = t
	return nil
}

// TraceFunc returns a new context and a function to finish the span.
func TraceFunc(ctx context.Context, operationName string, tags map[string]interface{}) (context.Context, func()) {
	span, ctx := opentracing.StartSpanFromContext(ctx, operationName)
	for k, v := range tags {
		span.SetTag(k, v)
	}
	return ctx, func() {
		span.Finish()
		if globalTracer != nil {
			globalTracer.logFunc("Span finished", "operation", operationName)
		}
	}
}
