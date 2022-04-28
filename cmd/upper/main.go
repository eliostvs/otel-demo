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
	libmath "github.com/username/otel-playground/internal/lib/math"
	"github.com/username/otel-playground/internal/lib/random"
	"github.com/username/otel-playground/internal/lib/telemetry"
	"github.com/username/otel-playground/internal/lib/web"
)

const (
	serviceName    = "upper"
	serviceVersion = "1.0.0"
)

var tracer trace.Tracer

var letters = []rune{
	'A',
	'B',
	'C',
	'D',
	'E',
	'F',
	'G',
	'H',
	'I',
	'J',
	'K',
	'L',
	'M',
	'N',
	'O',
	'P',
	'Q',
	'R',
	'S',
	'T',
	'U',
	'V',
	'W',
	'X',
	'Y',
	'Z',
}

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
	web.Handler(mux, "/", http.HandlerFunc(upperHandler))
	web.HealthCheckHandler(mux, serviceName, serviceVersion)

	if err := web.Server(port, mux, serviceName, web.FilterURLs{"/healthcheck"}); err != nil {
		log.Fatalf("failed to start server: %v\n", err)
	}
}

func upperHandler(w http.ResponseWriter, r *http.Request) {
	char := randomUpper(r.Context())

	char, err := processUpper(r.Context(), char)
	if err != nil {
		web.ServerErrorResponse(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, renderUpper(r.Context(), char))
}

func processUpper(ctx context.Context, char rune) (rune, error) {
	withAttr := trace.WithAttributes(attribute.String("char", string(char)))

	spctx, span := tracer.Start(ctx, "random_upper", withAttr, trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	span.AddEvent("processing_upper_char", withAttr)
	work(0.0001, 0.00005)

	// 1/100 calls is extra slow
	if rand.Float64() > 0.99 {
		span.AddEvent("extra_work", withAttr)
		work(0.0002, 0.0001)
	}

	// these chars are extra slow
	if collections.SliceContains(char, []rune{'Z', 'X', 'R'}) {
		if _, span := tracer.Start(ctx, "extra_process_upper", withAttr); span != nil {
			work(0.005, 0.0005)
			span.End()
		}
	}

	// these chars are extra slow and sometimes fail
	if collections.SliceContains(char, []rune{'Z', 'A', 'T'}) {
		if _, span := tracer.Start(spctx, "extra_extra_process_upper", withAttr); span != nil {
			defer span.End()
			work(0.0001, 0.00008)

			// fails 5% of the time
			if rand.Float64() > 0.95 {
				err := fmt.Errorf("failed to process '%c'", char)
				telemetry.RecordError(spctx, err)
				return -1, err
			}
		}
	}

	return char, nil
}

func randomUpper(ctx context.Context) rune {
	_, span := tracer.Start(ctx, "random_upper", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	// gets progressively slower throughout the hour
	work(float64(time.Now().Minute()/10000.0), 0.00001)

	char := random.Choice(letters)
	span.SetAttributes(attribute.String("char", string(char)))

	return char
}

func renderUpper(ctx context.Context, char rune) interface{} {
	_, span := tracer.Start(ctx, "render_upper", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	work(0.0002, 0.0001)

	return web.Envelope{"char": string(char)}
}

func work(mean, sigma float64) {
	time.Sleep(time.Duration(libmath.Max(0.0, random.Normalvariate(mean, sigma))))
}
