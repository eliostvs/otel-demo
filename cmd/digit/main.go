package main

import (
	"context"
	"flag"
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"

	"github.com/username/otel-playground/internal/lib/collections"
	"github.com/username/otel-playground/internal/lib/random"
	"github.com/username/otel-playground/internal/lib/telemetry"
	"github.com/username/otel-playground/internal/lib/web"
)

var digits = []rune{
	'0',
	'1',
	'2',
	'3',
	'4',
	'5',
	'6',
	'7',
	'8',
	'9',
}

const (
	serviceName    = "digit"
	serviceVersion = "1.0.0"
)

var tracer trace.Tracer

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	var port int
	flag.IntVar(&port, "port", 5000, "The port to listen on")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := telemetry.Configure(
		ctx,
		telemetry.WithServiceName(serviceName),
		telemetry.WithServiceVersion(serviceVersion),
	)
	if err != nil {
		log.Fatalf("failed to register tracer: %v\n", err)
	}
	defer func() {
		client.Shutdown(context.Background())
	}()

	tracer = otel.Tracer("main")

	mux := http.NewServeMux()
	web.Handler(mux, "/", http.HandlerFunc(digitHandler))
	web.HealthCheckHandler(mux, serviceName, serviceVersion)

	if err := web.Server(port, mux, serviceName, web.FilterURLs{"/healthcheck"}); err != nil {
		log.Fatalf("failed to start server: %v\n", err)
	}
}

func digitHandler(w http.ResponseWriter, r *http.Request) {
	char := randomDigit(r.Context())
	char = processDigit(r.Context(), char)
	web.WriteJSON(w, http.StatusOK, renderDigit(r.Context(), char))
}

func randomDigit(ctx context.Context) rune {
	bag := baggage.FromContext(ctx)
	_, span := tracer.Start(
		ctx,
		"random_digit",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("baggage.username", bag.Member("username").Value())),
	)
	defer span.End()

	work(0.0003, 0.0001)

	// slowness varies with the minute of the hour
	time.Sleep(time.Duration(math.Sin(float64(time.Now().Minute())) + 1.0))

	char := random.Choice(digits)
	span.SetAttributes(attribute.String("char", string(char)))

	return char
}

func processDigit(ctx context.Context, char rune) rune {
	opts := []trace.SpanStartOption{
		trace.WithAttributes(attribute.String("char", string(char))),
		trace.WithSpanKind(trace.SpanKindInternal),
	}

	attr := attribute.String("char", string(char))

	ctx, span := tracer.Start(ctx, "process_digit", opts...)
	defer span.End()

	work(0.0001, 0.00005)

	// 1/100 calls is extra slow when the digit is even
	// if random.random() > 0.99 and int(c) % 2 == 0:
	if rand.Float64() > 0.99 && char%2 == 0 {
		span.AddEvent("extra_work", trace.WithAttributes(attr))

		work(0.0002, 0.0001)
	}

	// these chars are extra slow
	if collections.SliceContains(char, []rune{'4', '5', '6'}) {
		if _, span := tracer.Start(ctx, "extra_process_digit", opts...); span != nil {
			work(0.005, 0.0005)
			span.End()
		}
	}

	return char
}

func renderDigit(ctx context.Context, char rune) web.Envelope {
	opts := []trace.SpanStartOption{
		trace.WithAttributes(attribute.String("char", string(char))),
		trace.WithSpanKind(trace.SpanKindInternal),
	}

	_, span := tracer.Start(ctx, "render_digit", opts...)
	defer span.End()

	work(0.0002, 0.0001)

	// every five minutes something goes wrong
	if time.Now().Minute()%5 == 0 {
		work(0.05, 0.005)
	}

	return web.Envelope{"char": string(char)}
}

func work(mean, sigma float64) {
	time.Sleep(time.Duration(random.Normalvariate(mean, sigma)))
}
