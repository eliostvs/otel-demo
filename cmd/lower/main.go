package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/username/otel-playground/internal/lib/environment"
	"github.com/username/otel-playground/internal/lib/random"
	"github.com/username/otel-playground/internal/lib/telemetry"
	"github.com/username/otel-playground/internal/lib/web"
)

const (
	serviceName    = "lower"
	serviceVersion = "1.0.0"
)

var letters = []rune{
	'a',
	'b',
	'c',
	'd',
	'e',
	'f',
	'g',
	'h',
	'i',
	'j',
	'k',
	'l',
	'm',
	'n',
	'o',
	'p',
	'q',
	'r',
	's',
	't',
	'u',
	'v',
	'w',
	'x',
	'y',
	'z',
}

var digitURL = environment.Get("DIGIT_URL", "http://digit:5000/")

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	var port int
	flag.IntVar(&port, "port", 5000, "The port to listen on")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	shutdownTracer, err := telemetry.RegisterTracer(ctx, serviceName, serviceVersion)
	if err != nil {
		log.Fatalf("failed to register tracer: %v\n", err)
	}
	defer func() {
		_ = shutdownTracer()
	}()

	mux := http.NewServeMux()
	web.Handler(mux, "/", http.HandlerFunc(lowerHandler))
	web.HealthCheckHandler(mux, serviceName, serviceVersion)

	if err := web.Server(port, mux); err != nil {
		log.Fatalf("failed to start server: %v\n", err)
	}
}

func lowerHandler(w http.ResponseWriter, r *http.Request) {
	char, err := randomLower(r.Context())
	if err != nil {
		web.ServerErrorResponse(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, web.Envelope{"char": string(char)})
}

func randomLower(ctx context.Context) (rune, error) {
	char := random.Choice(letters)

	_, err := getDigit(ctx, char)
	if err != nil {
		return -1, err
	}

	switch char {
	case 'z', 'x', 'r':
		work(ctx, 0.01, "extra_process_lower", char)
	case 'a', 't', 'y':
		work(ctx, 0.05, "extra_extra_process_lower", char)
	}

	return char, nil
}

func getDigit(ctx context.Context, char rune) (string, error) {
	spctx, span := telemetry.StartSpan(ctx, "digit", trace.WithAttributes(attribute.String("char", string(char))))
	defer span.End()

	var res struct {
		Char string `json:"char"`
	}

	if err := telemetry.GetJSON(spctx, digitURL, &res); err != nil {
		return "", fmt.Errorf("failed to fetch digit: %w", err)
	}

	return res.Char, nil
}

func work(ctx context.Context, await float64, spanName string, char rune) {
	_, span := telemetry.StartSpan(ctx, spanName, trace.WithAttributes(attribute.String("char", string(char))))
	time.Sleep(time.Duration(await))
	span.End()
}
