package web

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	metricGlobal "go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

const (
	instrumentationName    = "otel-playground"
	instrumentationVersion = "0.0.1"
)

type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterInterceptor) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func NewRequestCounterHandler(next http.Handler, filters FilterURLs) http.Handler {
	reqCounter, err := metricGlobal.Meter(
		instrumentationName,
		metric.WithInstrumentationVersion(instrumentationVersion),
	).
		SyncInt64().
		Counter("http.server.request_count")
	if err != nil {
		otel.Handle(err)
	}

	return &RequestCounterHandler{
		filters:    filters,
		reqCounter: reqCounter,
		next:       next,
	}
}

type RequestCounterHandler struct {
	reqCounter syncint64.Counter
	next       http.Handler
	filters    FilterURLs
}

func (h *RequestCounterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.filters.Use(r) {
		h.next.ServeHTTP(w, r)
		return
	}

	wi := &responseWriterInterceptor{
		statusCode:     http.StatusOK,
		ResponseWriter: w,
	}
	h.next.ServeHTTP(wi, r)

	h.reqCounter.Add(
		r.Context(),
		1,
		append(
			semconv.HTTPClientAttributesFromHTTPRequest(r),
			append(semconv.NetAttributesFromHTTPRequest("tcp", r), semconv.HTTPStatusCodeKey.Int(wi.statusCode))...,
		)...,
	)
}
