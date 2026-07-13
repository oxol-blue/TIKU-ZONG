package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oxol-blue/TIKU-ZONG/backend/internal/config"
)

func TestHealthz(t *testing.T) {
	router := NewRouter(config.Config{AppEnv: "test", AppName: "test-api"}, nil, nil, nil, nil, nil, nil)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected wildcard CORS header, got %q", got)
	}
}

func TestOptionsRequest(t *testing.T) {
	router := NewRouter(config.Config{AppEnv: "test"}, nil, nil, nil, nil, nil, nil)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/health", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Headers"); got == "" || !strings.Contains(got, "X-Admin-TOTP") {
		t.Fatalf("expected TOTP header to be allowed, got %q", got)
	}
}

func TestMetricsEndpoint(t *testing.T) {
	router := NewRouter(config.Config{AppEnv: "test"}, nil, nil, nil, nil, nil, nil)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/plain; version=0.0.4" {
		t.Fatalf("unexpected metrics content type %q", got)
	}
}

func TestReadinessWithoutDatabase(t *testing.T) {
	router := NewRouter(config.Config{AppEnv: "test"}, nil, nil, nil, nil, nil, nil)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", recorder.Code)
	}
}
