package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMetricsRecordsRequestAndLatency(t *testing.T) {
	t.Parallel()

	metrics := NewHTTPMetrics()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Pattern = "GET /health"
	rec := httptest.NewRecorder()

	Metrics(metrics)(next).ServeHTTP(rec, req)

	if err := testutil.GatherAndCompare(metrics.registry, strings.NewReader(`
# HELP task_manager_http_requests_total Total number of HTTP requests.
# TYPE task_manager_http_requests_total counter
task_manager_http_requests_total{method="GET",path="GET /health",status="200"} 1
`), "task_manager_http_requests_total"); err != nil {
		t.Fatalf("requests GatherAndCompare: %v", err)
	}

	families, err := metrics.registry.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	var foundDuration bool
	for _, family := range families {
		if family.GetName() != "task_manager_http_request_duration_seconds" {
			continue
		}
		foundDuration = true
		if len(family.GetMetric()) != 1 {
			t.Fatalf("duration metrics = %d, want 1", len(family.GetMetric()))
		}
		if family.GetMetric()[0].GetHistogram().GetSampleCount() != 1 {
			t.Fatalf("duration sample count = %d, want 1", family.GetMetric()[0].GetHistogram().GetSampleCount())
		}
	}
	if !foundDuration {
		t.Fatal("expected duration histogram metric")
	}
}

func TestMetricsRecordsErrors(t *testing.T) {
	t.Parallel()

	metrics := NewHTTPMetrics()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	req.Pattern = "GET /missing"
	rec := httptest.NewRecorder()

	Metrics(metrics)(next).ServeHTTP(rec, req)

	if err := testutil.GatherAndCompare(metrics.registry, strings.NewReader(`
# HELP task_manager_http_request_errors_total Total number of HTTP requests that returned an error status (>= 400).
# TYPE task_manager_http_request_errors_total counter
task_manager_http_request_errors_total{method="GET",path="GET /missing",status="404"} 1
`), "task_manager_http_request_errors_total"); err != nil {
		t.Fatalf("GatherAndCompare: %v", err)
	}
}

func TestMetricsSkipsMetricsEndpoint(t *testing.T) {
	t.Parallel()

	metrics := NewHTTPMetrics()
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	Metrics(metrics)(next).ServeHTTP(rec, req)

	if !called {
		t.Fatal("next handler should be called for /metrics")
	}

	families, err := metrics.registry.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	for _, family := range families {
		if family.GetName() == "task_manager_http_requests_total" {
			t.Fatal("expected /metrics request to be excluded from request metrics")
		}
	}
}

func TestMetricsHandlerExposesRegistry(t *testing.T) {
	t.Parallel()

	metrics := NewHTTPMetrics()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Pattern = "GET /health"
	rec := httptest.NewRecorder()
	Metrics(metrics)(next).ServeHTTP(rec, req)

	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsRec := httptest.NewRecorder()
	metrics.Handler().ServeHTTP(metricsRec, metricsReq)

	if metricsRec.Code != http.StatusOK {
		t.Fatalf("metrics status = %d, want %d", metricsRec.Code, http.StatusOK)
	}
	if !strings.Contains(metricsRec.Body.String(), "task_manager_http_requests_total") {
		t.Fatalf("metrics body = %q, want requests counter", metricsRec.Body.String())
	}
}

func TestMetricsNilMiddlewarePassesThrough(t *testing.T) {
	t.Parallel()

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	Metrics(nil)(next).ServeHTTP(rec, req)

	if !called {
		t.Fatal("next handler should be called when metrics is nil")
	}
}
