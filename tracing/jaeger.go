package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"

	"url-shortner/config"
	"url-shortner/log"
)

var DefaultTracer trace.Tracer

func NewJaegerTracer(ctx context.Context) error {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.DefaultConfig.Tracing.URL)))
	if err != nil {
		return err
	}

	traceProvider := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("urlShortener"),
		)),
	)
	go func() {
		<-ctx.Done()
		if err = traceProvider.Shutdown(ctx); err != nil {
			log.Errorf("Error shutting down tracer provider: %v", err)
		}
	}()
	tracer := traceProvider.Tracer("urlShortener")
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{},
		propagation.Baggage{}))
	DefaultTracer = tracer

	return err
}
