package config

import "os"

// Config contains process-level settings for the API service.
type Config struct {
	AppEnv            string
	AppName           string
	HTTPAddr          string
	MySQLDSN          string
	RedisAddr         string
	JWTSecret         string
	EncryptionSecret  string
	MySQLMaxOpenConns int
	MySQLMaxIdleConns int
	PublicBaseURL     string
}

// Load reads configuration from environment variables and applies local defaults.
func Load() Config {
	return Config{
		AppEnv:            envOrDefault("APP_ENV", "development"),
		AppName:           envOrDefault("APP_NAME", "tiku-zong-api"),
		HTTPAddr:          envOrDefault("HTTP_ADDR", ":8088"),
		MySQLDSN:          os.Getenv("MYSQL_DSN"),
		RedisAddr:         envOrDefault("REDIS_ADDR", "127.0.0.1:6379"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		EncryptionSecret:  envOrDefault("DATA_ENCRYPTION_SECRET", os.Getenv("JWT_SECRET")),
		MySQLMaxOpenConns: 10,
		MySQLMaxIdleConns: 5,
		PublicBaseURL:     envOrDefault("PUBLIC_BASE_URL", "http://localhost:8088"),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
