package web

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func Server(port int, handler http.Handler, serverName string, filters FilterURLs) error {
	var wg sync.WaitGroup

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: otelhttp.NewHandler(
			NewRequestCounterHandler(handler, filters),
			serverName,
			otelhttp.WithFilter(filters.Use),
		),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		log.Println("shutting down server")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			shutdownError <- err
		}

		log.Println("completing background tasks")

		wg.Wait()

		shutdownError <- nil
	}()

	log.Printf("port: %d", port)

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	if err := <-shutdownError; err != nil {
		return err
	}

	log.Printf("stopped server")

	return nil
}
