package telemetry

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"

	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

type localMachineDetector struct{}

func (l localMachineDetector) Detect(_ context.Context) (*sdkresource.Resource, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	addresses, err := net.LookupIP(hostname)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup hostname '%s':%w", hostname, err)
	}
	var ip net.IP
	for _, addr := range addresses {
		if ipv4 := addr.To4(); ipv4 != nil {
			ip = ipv4
			break
		}
	}
	if ip == nil {
		return nil, errors.New("failed to discover ip address")
	}

	return sdkresource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.HostNameKey.String(hostname),
		semconv.NetHostIPKey.String(ip.String()),
	), nil
}

func newResource(ctx context.Context, name, version string) (*sdkresource.Resource, error) {
	return sdkresource.New(
		ctx,
		sdkresource.WithSchemaURL(semconv.SchemaURL),
		sdkresource.WithFromEnv(),
		sdkresource.WithTelemetrySDK(),
		sdkresource.WithHost(),
		sdkresource.WithOS(),
		sdkresource.WithProcess(),
		sdkresource.WithAttributes(
			semconv.ServiceNameKey.String(name),
			semconv.ServiceVersionKey.String(version),
		),
		sdkresource.WithDetectors(&localMachineDetector{}),
	)
}
