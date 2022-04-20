package telemetry

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	hostMetrics "go.opentelemetry.io/contrib/instrumentation/host"
	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	metricGlobal "go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/nonrecording"
	sdkcontroller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	sdkprocessor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	sdkselector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"google.golang.org/grpc/encoding/gzip"
)

var (
	globalMeter = defaultMeter()
)

type meterHolder struct {
	m metric.Meter
}

func defaultMeter() *atomic.Value {
	v := &atomic.Value{}
	v.Store(meterHolder{m: nonrecording.NewNoopMeter()})
	return v
}

func SetMeter(meter metric.Meter) {
	globalMeter.Store(meterHolder{meter})
}

func Meter() metric.Meter {
	return globalMeter.Load().(meterHolder).m
}

func RegisterMeter(ctx context.Context, name, version string) (func() error, error) {
	resource, err := newResource(ctx, name, version)
	if err != nil {
		return nil, fmt.Errorf("failed to create otel resource: %w", err)
	}

	exporter, err := newMetricExporter(ctx)
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
		sdkcontroller.WithCollectPeriod(30*time.Second),
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

	return func() error {
		ctx := context.Background()
		_ = pusher.Stop(ctx)
		return exporter.Shutdown(ctx)
	}, nil
}

func newMetricExporter(_ context.Context) (*otlpmetric.Exporter, error) {
	return otlpmetric.New(
		context.Background(),
		otlpmetricgrpc.NewClient(
			otlpmetricgrpc.WithCompressor(gzip.Name),
			otlpmetricgrpc.WithInsecure(),
		),
	)
}
