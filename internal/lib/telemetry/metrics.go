package telemetry

//
// import (
// 	"context"
// 	"fmt"
//
// 	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
// 	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
// 	"go.opentelemetry.io/otel/metric"
// 	"go.opentelemetry.io/otel/metric/global"
// 	sdkcontroller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
// 	"go.opentelemetry.io/otel/sdk/metric/export"
// 	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
// 	sdkprocessor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
// 	sdkselector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
// 	sdkresource "go.opentelemetry.io/otel/sdk/resource"
// 	"google.golang.org/grpc/encoding/gzip"
// )
//
// func RegisterMetric(ctx context.Context, name, version string) (metric.Meter, error) {
// 	resource, err := newResource(ctx, name, version)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create otel resource: %w", err)
// 	}
//
// 	exporter, err := newMetricExporter(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	provider, err := newMetricProvider(ctx, resource, exporter)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	global.SetMeterProvider(provider)
// 	return global.MeterProvider().Meter(name, metric.WithInstrumentationVersion(version)), nil
// }
//
// func newMetricExporter(ctx context.Context) (*otlpmetric.Exporter, error) {
// 	exporter, err := otlpmetric.New(
// 		ctx,
// 		newMetricClient(),
// 		// what is this for?
// 		otlpmetric.WithMetricAggregationTemporalitySelector(aggregation.StatelessTemporalitySelector()),
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create exporter: %w", err)
// 	}
//
// 	return exporter, nil
// }
//
// func newMetricClient() otlpmetric.Client {
// 	return otlpmetricgrpc.NewClient(
// 		otlpmetricgrpc.WithCompressor(gzip.Name),
// 		otlpmetricgrpc.WithInsecure(),
// 	)
// }
//
// func newMetricProvider(ctx context.Context, resource *sdkresource.Resource, exporter export.Exporter) (
// 	*sdkcontroller.Controller,
// 	error,
// ) {
// 	controller := sdkcontroller.New(
// 		sdkprocessor.NewFactory(
// 			// what is this for?
// 			sdkselector.NewWithHistogramDistribution(),
// 			// what is this for?
// 			aggregation.StatelessTemporalitySelector(),
// 		),
// 		sdkcontroller.WithExporter(exporter),
// 		sdkcontroller.WithResource(resource),
// 	)
//
// 	if err := controller.Start(ctx); err != nil {
// 		return nil, fmt.Errorf("failed to start provider: %w", err)
// 	}
//
// 	return controller, nil
// }
