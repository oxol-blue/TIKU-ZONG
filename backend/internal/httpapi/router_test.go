package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oxol-blue/TIKU-ZONG/backend/internal/config"
)

func TestHealthz(t *testing.T) {
	router := NewRouter(config.Config{AppEnv: "test", AppName: "test-api"}, nil, nil, nil, nil)
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
	router := NewRouter(config.Config{AppEnv: "test"}, nil, nil, nil, nil)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/health", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", recorder.Code)
	}
}
