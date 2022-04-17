package telemetry

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/encoding/gzip"
)

var (
	globalTracer        = defaultTracer()
	setGlobalTracerOnce sync.Once
)

type tracerHolder struct {
	t trace.Tracer
}

func defaultTracer() *atomic.Value {
	v := &atomic.Value{}
	v.Store(tracerHolder{t: trace.NewNoopTracerProvider().Tracer("")})
	return v
}

func setGlobalTracer(t trace.Tracer) {
	setGlobalTracerOnce.Do(
		func() {
			globalTracer.Store(tracerHolder{t})
		},
	)
}

func Tracer() trace.Tracer {
	return globalTracer.Load().(tracerHolder).t
}

func RegisterTracer(ctx context.Context, name, version string) error {
	resource, err := newResource(ctx, name, version)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	exporter, err := NewOTLPTraceExporter(ctx)
	if err != nil {
		return fmt.Errorf("failed to create exporter: %w", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())),
		sdktrace.WithResource(resource),
		sdktrace.WithBatcher(exporter),
	)

	otel.SetTracerProvider(provider)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	setGlobalTracer(
		provider.Tracer(
			name,
			trace.WithInstrumentationVersion(version),
		),
	)

	return nil
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

func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	opts = append(opts, trace.WithSpanKind(trace.SpanKindInternal))
	spctx, span := Tracer().Start(ctx, name, opts...)
	return spctx, span
}

func RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err, trace.WithStackTrace(true))
	span.SetStatus(codes.Error, err.Error())
}
