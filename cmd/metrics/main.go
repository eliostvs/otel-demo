package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"

	"github.com/username/otel-playground/internal/lib/telemetry"
)

var meter metric.Meter

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown, err := telemetry.RegisterMeter(ctx, "foo", "0.0.1")
	if err != nil {
		log.Fatal(err)

	}
	defer func() {
		_ = shutdown()
	}()

	meter = global.MeterProvider().Meter("app_test")
	go counterMetric(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit
}

func counterMetric(ctx context.Context) {
	counter, err := meter.SyncInt64().Counter(
		"foo.prefix.counter",
		instrument.WithUnit("1"),
		instrument.WithDescription("TODO"),
	)
	if err != nil {
		log.Fatal(err)
	}

	for {
		log.Println("send counter")
		counter.Add(ctx, 1)
		time.Sleep(time.Millisecond)
	}
}
