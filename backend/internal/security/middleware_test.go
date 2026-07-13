package security

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRateLimitAndBlacklist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(NewMiddleware(Config{RateLimitPerMinute: 1, Blacklist: []string{"10.0.0.2"}}))
	router.GET("/api/v1/search", func(c *gin.Context) { c.Status(http.StatusOK) })
	request := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=x", nil)
	request.RemoteAddr = "10.0.0.1:1234"
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("first request status = %d", recorder.Code)
	}
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("second request status = %d", recorder.Code)
	}
	blocked := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=x", nil)
	blocked.RemoteAddr = "10.0.0.2:1234"
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, blocked)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("blocked request status = %d", recorder.Code)
	}
}
