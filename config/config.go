package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string
	AppEnv  string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	RedisHost     string
	RedisPort     string
	RedisPassword string

	JWTSecret      string
	JWTExpiryHours int

	RateLimitCapacity   int
	RateLimitRefillRate int

	HealthCheckIntervalSeconds int
	HealthCheckWorkerPoolSize  int
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment directly")
	}

	return &Config{
		AppPort: getEnv("APP_PORT", "8080"),
		AppEnv:  getEnv("APP_ENV", "development"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "gosnip"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "gosnip_db"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		JWTSecret:      getEnv("JWT_SECRET", "changeme"),
		JWTExpiryHours: getEnvInt("JWT_EXPIRY_HOURS", 24),

		RateLimitCapacity:   getEnvInt("RATE_LIMIT_CAPACITY", 10),
		RateLimitRefillRate: getEnvInt("RATE_LIMIT_REFILL_RATE", 1),

		HealthCheckIntervalSeconds: getEnvInt("HEALTH_CHECK_INTERVAL_SECONDS", 300),
		HealthCheckWorkerPoolSize:  getEnvInt("HEALTH_CHECK_WORKER_POOL_SIZE", 10),
	}
}

// read env variable with default value
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Invalid int value for %s, using default: %d", key, fallback)
		return fallback
	}
	return i
}
