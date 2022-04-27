package telemetry

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
)

type setupFunc func(context.Context, *resource.Resource) (func(context.Context) error, error)

type Client struct {
	config        Config
	shutdownFuncs []func(context.Context) error
}

func Configure(ctx context.Context, opts ...Option) (Client, error) {
	cfg := newConfig(opts...)

	res, err := newResource(ctx, cfg)
	if err != nil {
		return Client{}, err
	}
	c := Client{config: cfg}

	if cfg.errorHandler != nil {
		otel.SetErrorHandler(cfg.errorHandler)
	}

	for _, setup := range []setupFunc{configureMetrics, configureTracing} {
		shutdown, err := setup(ctx, res)
		if err != nil {
			continue
		}

		if shutdown != nil {
			c.shutdownFuncs = append(c.shutdownFuncs, shutdown)
		}
	}

	return c, nil
}

func (c Client) Shutdown(ctx context.Context) {
	for _, shutdown := range c.shutdownFuncs {
		if err := shutdown(ctx); err != nil {
			log.Printf("failed to stop exporters: %s\n", err)
		}
	}
}
