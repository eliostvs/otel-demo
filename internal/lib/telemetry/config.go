package telemetry

import (
	"go.opentelemetry.io/otel"
)

type (
	Config struct {
		serviceName    string
		serviceVersion string
		metricsEnabled bool
		tracingEnabled bool
		errorHandler   otel.ErrorHandler
	}

	Option func(*Config)
)

func newConfig(opts ...Option) Config {
	var defaultOpts []Option

	c := Config{
		tracingEnabled: true,
		metricsEnabled: true,
		errorHandler:   errorHandler{},
	}

	for _, opt := range append(defaultOpts, opts...) {
		opt(&c)
	}

	return c
}

// WithErrorHandler configures a global error handler to be used throughout an OpenTelemetry instrumented project.
func WithErrorHandler(handler otel.ErrorHandler) Option {
	return func(c *Config) {
		c.errorHandler = handler
	}
}

// WithServiceName configures a "service.name" resource label
func WithServiceName(name string) Option {
	return func(c *Config) {
		c.serviceName = name
	}
}

// WithServiceVersion configures a "service.version" resource label
func WithServiceVersion(version string) Option {
	return func(c *Config) {
		c.serviceVersion = version
	}
}

func WithMetricsEnabled(enabled bool) Option {
	return func(c *Config) {
		c.metricsEnabled = enabled
	}
}

func WithTracingEnabled(enabled bool) Option {
	return func(c *Config) {
		c.tracingEnabled = enabled
	}
}
