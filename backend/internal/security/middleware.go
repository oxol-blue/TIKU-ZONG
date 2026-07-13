package security

import (
	"bufio"
	"fmt"
	"net"
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
	RedisAddr          string
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
	redisAddr string
}

func NewMiddleware(config Config) gin.HandlerFunc {
	middleware := &Middleware{config: config, blacklist: toSet(config.Blacklist), whitelist: toSet(config.Whitelist), counters: make(map[string]counter), redisAddr: strings.TrimSpace(config.RedisAddr)}
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
		_, allowed := m.increment(ip, window)
		if !allowed {
			c.Header("Retry-After", strconv.Itoa(60-window.Second()))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"code": "RATE_LIMITED", "message": "too many search requests"})
			return
		}
	}
	c.Next()
}

func (m *Middleware) increment(ip string, window time.Time) (int, bool) {
	if m.redisAddr != "" {
		if count, ok := redisIncrement(m.redisAddr, "tiku:rate:"+ip+":"+strconv.FormatInt(window.Unix(), 10)); ok {
			return count, count <= m.config.RateLimitPerMinute
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	value := m.counters[ip]
	if !value.window.Equal(window) {
		value = counter{window: window}
	}
	value.count++
	m.counters[ip] = value
	return value.count, value.count <= m.config.RateLimitPerMinute
}

func redisIncrement(address, key string) (int, bool) {
	connection, err := net.DialTimeout("tcp", address, 150*time.Millisecond)
	if err != nil {
		return 0, false
	}
	defer connection.Close()
	_ = connection.SetDeadline(time.Now().Add(250 * time.Millisecond))
	command := respCommand("EVAL", "local c=redis.call('INCR',KEYS[1]);if c==1 then redis.call('EXPIRE',KEYS[1],ARGV[1]);end;return c", "1", key, "60")
	if _, err := connection.Write([]byte(command)); err != nil {
		return 0, false
	}
	line, err := bufio.NewReader(connection).ReadString('\n')
	if err != nil || len(line) < 2 || line[0] != ':' {
		return 0, false
	}
	count, err := strconv.Atoi(strings.TrimSpace(line[1:]))
	return count, err == nil
}

func respCommand(values ...string) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "*%d\r\n", len(values))
	for _, value := range values {
		fmt.Fprintf(&builder, "$%d\r\n%s\r\n", len([]byte(value)), value)
	}
	return builder.String()
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
