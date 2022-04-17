package telemetry

import (
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

func NewTransport() *otelhttp.Transport {
	return otelhttp.NewTransport(
		nil,
		otelhttp.WithPropagators(otel.GetTextMapPropagator()),
	)
}
