package telemetry

import (
	"log"
)

type errorHandler struct {
}

func (e errorHandler) Handle(err error) {
	log.Printf("otel error handler: %v\n", err)
}
