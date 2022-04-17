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
)

func Server(port int, mux http.Handler) error {
	var wg sync.WaitGroup

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
		// ErrorLog:     log.New(app.logger, "", 0),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		// app.logger.PrintInfo("shutting down server", map[string]string{
		// 	"signal": s.String(),
		// })

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			shutdownError <- err
		}

		// app.logger.PrintInfo("completing background tasks", map[string]string{
		// 	"addr": srv.Addr,
		// })

		wg.Wait()

		shutdownError <- nil
	}()

	log.Printf("port: %d", port)
	// app.logger.PrintInfo(
	// 	"starting server", map[string]string{
	// 		"addr": srv.Addr,
	// 		"env":  app.config.env,
	// 	},
	// )

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	if err := <-shutdownError; err != nil {
		return err
	}

	// app.logger.PrintInfo(
	// 	"stopped server", map[string]string{
	// 		"addr": srv.Addr,
	// 	},
	// )

	return nil
}
