package web

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"

	"github.com/username/otel-playground/internal/lib/collections"
	libjson "github.com/username/otel-playground/internal/lib/json"
	"github.com/username/otel-playground/internal/lib/telemetry"
)

func NewTransport() *otelhttp.Transport {
	return otelhttp.NewTransport(
		nil,
		otelhttp.WithPropagators(otel.GetTextMapPropagator()),
	)
}

var successfulStatuses = []int{
	http.StatusOK,
	http.StatusCreated,
	http.StatusAccepted,
	http.StatusNonAuthoritativeInfo,
	http.StatusNoContent,
}

type ResponseError http.Response

// Error fulfills the error interface.
func (se *ResponseError) Error() string {
	return fmt.Sprintf("response error for %s", se.Request.URL.Redacted())
}

// GetJSON fetch the given url and try to decode the response as json
// any error will be record to the trace
func GetJSON(ctx context.Context, url string, dst interface{}) (err error) {
	defer func() {
		telemetry.RecordResult(ctx, err)
	}()

	res, err := otelhttp.Get(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to fetch '%s': %w", url, err)
	}
	defer res.Body.Close()

	if !collections.SliceContains(res.StatusCode, successfulStatuses) {
		return fmt.Errorf("%w: unexpected status: %d", (*ResponseError)(res), res.StatusCode)
	}

	if err := libjson.Decode(res.Body, dst); err != nil {
		return fmt.Errorf("failed to decode json body: %w", err)
	}

	return nil
}
