package telemetry

import (
	"context"
	"fmt"

	hostMetrics "go.opentelemetry.io/contrib/instrumentation/host"
	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	metricGlobal "go.opentelemetry.io/otel/metric/global"
	sdkcontroller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	sdkprocessor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	sdkselector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc/encoding/gzip"
)

func configureMetrics(ctx context.Context, resource *resource.Resource) (func(context.Context) error, error) {
	exporter, err := newOTLPMetricExporter(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	pusher := sdkcontroller.New(
		sdkprocessor.NewFactory(
			sdkselector.NewWithInexpensiveDistribution(),
			exporter,
		),
		sdkcontroller.WithExporter(exporter),
		sdkcontroller.WithResource(resource),
	)

	if err := pusher.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start metric provider: %w", err)
	}

	if err := runtimemetrics.Start(runtimemetrics.WithMeterProvider(pusher)); err != nil {
		return nil, fmt.Errorf("runtimemetrics.Start failed: %s", err)
	}

	if err = hostMetrics.Start(hostMetrics.WithMeterProvider(pusher)); err != nil {
		return nil, fmt.Errorf("failed to start host metrics: %v", err)
	}

	metricGlobal.SetMeterProvider(pusher)

	return func(ctx context.Context) (lastErr error) {
		if err := pusher.Stop(ctx); err != nil {
			lastErr = err
		}

		if err := exporter.Shutdown(ctx); err != nil {
			lastErr = err
		}

		return lastErr
	}, nil
}

func newOTLPMetricExporter(ctx context.Context) (*otlpmetric.Exporter, error) {
	return otlpmetric.New(
		ctx,
		otlpmetricgrpc.NewClient(
			otlpmetricgrpc.WithCompressor(gzip.Name),
			otlpmetricgrpc.WithInsecure(),
		),
	)
}
