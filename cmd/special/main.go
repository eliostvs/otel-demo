package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/username/otel-playground/internal/lib/collections"
	"github.com/username/otel-playground/internal/lib/random"
	"github.com/username/otel-playground/internal/lib/telemetry"
	"github.com/username/otel-playground/internal/lib/web"
)

const (
	serviceName    = "special"
	serviceVersion = "1.0.0"
)

var characters = []rune{
	'!',
	'@',
	'#',
	'$',
	'%',
	'^',
	'&',
	'*',
	'<',
	'>',
	',',
	'.',
	':',
	';',
	'?',
	'/',
	'+',
	'=',
	'{',
	'}',
	'[',
	']',
	'-',
	'_',
	'\\',
	'|',
	'~',
	'`',
}

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
	web.Handler(mux, "/", http.HandlerFunc(specialHandler))
	web.HealthCheckHandler(mux, serviceName, serviceVersion)

	if err := web.Server(port, mux, serviceName, web.FilterURLs{"/healthcheck"}); err != nil {
		log.Fatalf("failed to start server: %v\n", err)
	}
}

func specialHandler(w http.ResponseWriter, r *http.Request) {
	char := randomSpecial(r.Context())

	char, err := processSpecial(r.Context(), char)
	if err != nil {
		web.ServerErrorResponse(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, renderSpecial(r.Context(), char))
}

func randomSpecial(ctx context.Context) rune {
	_, span := tracer.Start(ctx, "random_special", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	work(0.0003, 0.0001)
	char := random.Choice(characters)
	span.SetAttributes(attribute.String("char", string(char)))

	return char
}

func processSpecial(ctx context.Context, char rune) (rune, error) {
	opts := []trace.SpanStartOption{
		trace.WithAttributes(attribute.String("char", string(char))),
		trace.WithSpanKind(trace.SpanKindInternal),
	}

	spctx, span := tracer.Start(ctx, "process_special", opts...)
	defer span.End()

	work(0.0001, 0.00005)

	// these chars are extra slow
	if collections.SliceContains(char, []rune{'$', '@', '#', '?', '%'}) {
		if _, span := tracer.Start(spctx, "extra_process_special", opts...); span != nil {
			work(0.005, 0.0005)
			span.End()
		}
	}

	// these chars fail 5% of the time
	if collections.SliceContains(char, []rune{'!', '@', '?'}) && rand.Float64() > 0.95 {
		err := fmt.Errorf("Failed to process '%c' ", char)
		telemetry.RecordError(spctx, err)
		return -1, err
	}

	return char, nil
}

func renderSpecial(ctx context.Context, char rune) interface{} {
	attr := attribute.String("char", string(char))

	_, span := tracer.Start(
		ctx, "render_special", trace.WithAttributes(attr), trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	span.AddEvent("processing special char", trace.WithAttributes(attr))

	work(0.0002, 0.0001)

	return web.Envelope{"char": string(char)}
}

// work simulates work being done.
func work(mean, sigma float64) {
	time.Sleep(time.Duration(random.Normalvariate(mean, sigma)))
}
