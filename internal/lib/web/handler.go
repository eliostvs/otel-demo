package web

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func Handler(mux *http.ServeMux, route string, handler http.Handler) {
	mux.Handle(route, otelhttp.WithRouteTag(route, handler))
}

func HealthCheckHandler(mux *http.ServeMux, service, version string) {
	mux.HandleFunc(
		"/healthcheck", func(w http.ResponseWriter, r *http.Request) {
			data := Envelope{
				"service": service,
				"version": version,
				"status":  "ok",
			}
			WriteJSON(w, http.StatusOK, data)
		},
	)
}
