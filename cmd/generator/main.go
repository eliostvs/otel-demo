package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/username/otel-playground/internal/lib/environment"
	"github.com/username/otel-playground/internal/lib/random"
	"github.com/username/otel-playground/internal/lib/telemetry"
	"github.com/username/otel-playground/internal/lib/web"
)

const (
	serviceName    = "generator"
	serviceVersion = "1.0.0"
)

var tracer trace.Tracer

type generator struct {
	name, url string
}

var generators = []generator{
	{"generator.uppers", environment.Get("UPPER_URL", "http://upper:5000/")},
	{"generator.lowers", environment.Get("LOWER_URL", "http://lower:5000/")},
	{"generator.digits", environment.Get("DIGIT_URL", "http://digit:5000/")},
	{"generator.specials", environment.Get("SPECIAL_URL", "http://special:5000/")},
}

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	var port int
	flag.IntVar(&port, "port", 5000, "The port to listen on")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
	web.Handler(mux, "/", http.HandlerFunc(generatorHandler))
	web.HealthCheckHandler(mux, serviceName, serviceVersion)

	if err := web.Server(port, mux); err != nil {
		log.Fatalf("failed to start server: %v\n", err)
	}
}

func generatorHandler(w http.ResponseWriter, r *http.Request) {
	password, err := generate(r.Context())
	if err != nil {
		web.ServerErrorResponse(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, web.Envelope{"password": password})
}

func generate(ctx context.Context) (string, error) {
	spctx, span := tracer.Start(ctx, "generator.generate", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	var password []string
	span.AddEvent("selecting_password_length")
	work(0.00001, 0.00001)
	passwordLength := random.NumberInRange(8, 25)
	span.SetAttributes(attribute.Int("password.length", passwordLength))

	i := 1
	for len(password) < passwordLength {
		log.Printf("generate_loop_%d\n", i)
		span.AddEvent(fmt.Sprintf("generate_loop_%d", i), trace.WithAttributes(attribute.Int("iteration", i)))

		for _, gen := range generators {
			chars, err := getChars(spctx, gen.name, gen.url)
			if err != nil {
				return "", err
			}
			password = append(password, chars...)
		}
		i++
	}
	span.AddEvent("shuffling_password")
	rand.Shuffle(
		len(password), func(i, j int) {
			password[i], password[j] = password[j], password[i]
		},
	)

	if len(password) > passwordLength {
		span.AddEvent("trimming_password", trace.WithAttributes(attribute.Int("password.length", passwordLength)))
		password = password[0:passwordLength]
	}

	return strings.Join(password, ""), nil
}

func getChars(ctx context.Context, spanName, url string) ([]string, error) {
	spctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	var x []string
	for i := 0; i < random.NumberInRange(0, 3); i++ {
		log.Printf("%s, iteration_loop_%d\n", spanName, i)
		span.AddEvent(fmt.Sprintf("iteration_%d", i), trace.WithAttributes(attribute.Int("iteration", i)))

		var resp struct {
			Char string `json:"char"`
		}

		if err := telemetry.GetJSON(spctx, url, &resp); err != nil {
			return nil, fmt.Errorf("failed to fetch url '%s': %w", url, err)
		}

		x = append(x, resp.Char)
	}

	if len(x) == 0 {
		span.AddEvent(spanName + ".ignored")
	}

	return x, nil
}

func work(mean, sigma float64) {
	time.Sleep(time.Duration(random.Normalvariate(mean, sigma)))
}
