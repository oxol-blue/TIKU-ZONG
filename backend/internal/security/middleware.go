package security

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Config struct {
	RateLimitPerMinute int
	Blacklist          []string
	Whitelist          []string
}

type counter struct {
	window time.Time
	count  int
}

type Middleware struct {
	config    Config
	blacklist map[string]struct{}
	whitelist map[string]struct{}
	mu        sync.Mutex
	counters  map[string]counter
}

func NewMiddleware(config Config) gin.HandlerFunc {
	middleware := &Middleware{config: config, blacklist: toSet(config.Blacklist), whitelist: toSet(config.Whitelist), counters: make(map[string]counter)}
	return middleware.Handle
}

func (m *Middleware) Handle(c *gin.Context) {
	ip := c.ClientIP()
	if len(m.whitelist) > 0 {
		if _, ok := m.whitelist[ip]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": "IP_NOT_ALLOWED", "message": "client IP is not allowlisted"})
			return
		}
	}
	if _, ok := m.blacklist[ip]; ok {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": "IP_BLOCKED", "message": "client IP is blocked"})
		return
	}
	if m.config.RateLimitPerMinute > 0 && rateLimitedPath(c.Request.URL.Path) {
		window := time.Now().UTC().Truncate(time.Minute)
		m.mu.Lock()
		value := m.counters[ip]
		if !value.window.Equal(window) {
			value = counter{window: window}
		}
		value.count++
		m.counters[ip] = value
		allowed := value.count <= m.config.RateLimitPerMinute
		m.mu.Unlock()
		if !allowed {
			c.Header("Retry-After", strconv.Itoa(60-window.Second()))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"code": "RATE_LIMITED", "message": "too many search requests"})
			return
		}
	}
	c.Next()
}

func rateLimitedPath(path string) bool {
	return strings.HasPrefix(path, "/api/v1/search") || strings.HasPrefix(path, "/api/ocs/search")
}

func toSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result[value] = struct{}{}
		}
	}
	return result
}
