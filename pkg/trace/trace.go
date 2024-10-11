package trace

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
)

var globalTracer *Tracer

type Tracer struct {
	logger *zap.Logger
}

type TracerOption func(*Tracer)

func WithLogger(logger *zap.Logger) TracerOption {
	return func(t *Tracer) {
		t.logger = logger
	}
}

// Init initializes the global tracer and returns an error if initialization fails
func Init(endpoint, serviceName string, options ...TracerOption) error {
	t := &Tracer{
		logger: zap.NewNop(), // Use a no-op logger by default
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
		return fmt.Errorf("failed to init tracing: %w", err)
	}

	t.logger.Debug("Tracing initialized", zap.String("endpoint", endpoint), zap.String("service", serviceName))
	globalTracer = t
	return nil
}

// TraceFunc returns a new context and a function to finish the span.
func TraceFunc(ctx context.Context, operationName string, tags map[string]interface{}) (context.Context, func()) {
	span, ctx := opentracing.StartSpanFromContext(ctx, operationName)
	for k, v := range tags {
		span.SetTag(k, v)
	}
	return ctx, func() { span.Finish() }
}
