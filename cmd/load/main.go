package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/username/otel-playground/internal/lib/environment"
	"github.com/username/otel-playground/internal/lib/random"
	"github.com/username/otel-playground/internal/lib/web"
)

func main() {
	url := environment.Get("GENERATOR_URL", "http://generator:5000/")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	for {
		select {
		case <-quit:
			os.Exit(0)
		case <-time.After(time.Duration(random.NumberInRange(351, 795)) * time.Millisecond):
			getPassword(url)
		}
	}
}

func getPassword(url string) {
	var res struct {
		Password string `json:"password"`
		Cause    string `json:"cause"`
	}

	if err := web.GetJSON(url, &res); err != nil {
		log.Printf("failed to get password: %v\n", err)
		return
	}

	log.Printf("got password '%s'\n", res.Password)
}
