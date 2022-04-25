package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/encoding/gzip"
)

func ConfigureTracing(ctx context.Context, resource *resource.Resource) (func(context.Context) error, error) {
	exporter, err := NewOTLPTraceExporter(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())),
		sdktrace.WithResource(resource),
		sdktrace.WithBatcher(exporter),
	)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	otel.SetTracerProvider(provider)

	return func(ctx context.Context) error {
		return provider.Shutdown(ctx)
	}, nil
}

func NewOTLPTraceExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	return otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithCompressor(gzip.Name),
			otlptracegrpc.WithInsecure(),
		),
	)
}

func Span(ctx context.Context, tracer trace.Tracer, name string, opts ...trace.SpanStartOption) (
	context.Context,
	trace.Span,
) {
	opts = append(opts, trace.WithSpanKind(trace.SpanKindInternal))
	return tracer.Start(ctx, name, opts...)
}

func RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err, trace.WithStackTrace(true))
	span.SetStatus(codes.Error, err.Error())
}

func RecordResult(ctx context.Context, err error) {
	if err != nil {
		RecordError(ctx, err)
		return
	}
	span := trace.SpanFromContext(ctx)
	span.SetStatus(codes.Ok, codes.Ok.String())
}
