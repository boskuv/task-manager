package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	metricNamespace = "task_manager"
)

// HTTPMetrics holds Prometheus collectors for HTTP traffic.
type HTTPMetrics struct {
	RequestsTotal *prometheus.CounterVec
	ErrorsTotal   *prometheus.CounterVec
	Duration      *prometheus.HistogramVec
	registry      *prometheus.Registry
}

// NewHTTPMetrics registers HTTP collectors in a dedicated Prometheus registry.
func NewHTTPMetrics() *HTTPMetrics {
	registry := prometheus.NewRegistry()

	requestsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricNamespace,
		Name:      "http_requests_total",
		Help:      "Total number of HTTP requests.",
	}, []string{"method", "path", "status"})

	errorsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricNamespace,
		Name:      "http_request_errors_total",
		Help:      "Total number of HTTP requests that returned an error status (>= 400).",
	}, []string{"method", "path", "status"})

	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metricNamespace,
		Name:      "http_request_duration_seconds",
		Help:      "HTTP request latency in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "path"})

	registry.MustRegister(requestsTotal, errorsTotal, duration)

	return &HTTPMetrics{
		RequestsTotal: requestsTotal,
		ErrorsTotal:   errorsTotal,
		Duration:      duration,
		registry:      registry,
	}
}

// Handler exposes collected metrics for Prometheus scraping.
func (m *HTTPMetrics) Handler() http.Handler {
	if m == nil {
		return http.NotFoundHandler()
	}
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// Metrics records request counts, error counts, and latency for each HTTP call.
func Metrics(m *HTTPMetrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if m == nil || r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(recorder, r)

			path := routePattern(r)
			status := strconv.Itoa(recorder.status)
			labels := prometheus.Labels{
				"method": r.Method,
				"path":   path,
				"status": status,
			}

			m.RequestsTotal.With(labels).Inc()
			if recorder.status >= http.StatusBadRequest {
				m.ErrorsTotal.With(labels).Inc()
			}
			m.Duration.With(prometheus.Labels{
				"method": r.Method,
				"path":   path,
			}).Observe(time.Since(start).Seconds())
		})
	}
}

func routePattern(r *http.Request) string {
	if r.Pattern != "" {
		return r.Pattern
	}
	if r.URL != nil && r.URL.Path != "" {
		return r.URL.Path
	}
	return "unknown"
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}
