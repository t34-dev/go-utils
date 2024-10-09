package trace

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
)

// Init initializes tracing with jaeger
// example endpoint: localhost:6831
func Init(endpoint string, logger *zap.Logger, serviceName string) {
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
		logger.Fatal("failed to init tracing", zap.Error(err))
	}
}

// TraceFunc returns a new context and a function to finish the span.
func TraceFunc(ctx context.Context, operationName string, tags map[string]interface{}) (context.Context, func()) {
	span, ctx := opentracing.StartSpanFromContext(ctx, operationName)
	for k, v := range tags {
		span.SetTag(k, v)
	}
	return ctx, func() { span.Finish() }
}
