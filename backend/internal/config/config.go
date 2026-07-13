package config

import (
	"os"
	"strconv"
	"strings"
)

// Config contains process-level settings for the API service.
type Config struct {
	AppEnv                string
	AppName               string
	HTTPAddr              string
	MySQLDSN              string
	RedisAddr             string
	JWTSecret             string
	EncryptionSecret      string
	MySQLMaxOpenConns     int
	MySQLMaxIdleConns     int
	PublicBaseURL         string
	APIRateLimitPerMinute int
	IPBlacklist           []string
	IPWhitelist           []string
	AdminTOTPSecret       string
	AnswerMergeRule       string
}

// Load reads configuration from environment variables and applies local defaults.
func Load() Config {
	return Config{
		AppEnv:                envOrDefault("APP_ENV", "development"),
		AppName:               envOrDefault("APP_NAME", "tiku-zong-api"),
		HTTPAddr:              envOrDefault("HTTP_ADDR", ":8088"),
		MySQLDSN:              os.Getenv("MYSQL_DSN"),
		RedisAddr:             envOrDefault("REDIS_ADDR", "127.0.0.1:6379"),
		JWTSecret:             os.Getenv("JWT_SECRET"),
		EncryptionSecret:      envOrDefault("DATA_ENCRYPTION_SECRET", os.Getenv("JWT_SECRET")),
		MySQLMaxOpenConns:     10,
		MySQLMaxIdleConns:     5,
		PublicBaseURL:         envOrDefault("PUBLIC_BASE_URL", "http://localhost:8088"),
		APIRateLimitPerMinute: envIntOrDefault("API_RATE_LIMIT_PER_MINUTE", 120),
		IPBlacklist:           splitList(os.Getenv("IP_BLACKLIST")),
		IPWhitelist:           splitList(os.Getenv("IP_WHITELIST")),
		AdminTOTPSecret:       os.Getenv("ADMIN_TOTP_SECRET"),
		AnswerMergeRule:       envOrDefault("ANSWER_MERGE_RULE", "priority"),
	}
}

func envIntOrDefault(key string, fallback int) int {
	value := os.Getenv(key)
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return fallback
	}
	return parsed
}

func splitList(value string) []string { return strings.Split(value, ",") }

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
