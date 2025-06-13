package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost                  string
	DBPort                  string
	DBUsername              string
	DBPassword              string
	DBDatabase              string
	Port                    string
	LogLevel                string
	WorkerEnabled           bool
	WorkerUpdateInterval    int
	WorkerDiscoveryInterval int
	FanslyAPIRateLimit      int
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		DBHost:                  getEnv("DB_HOST", "localhost"),
		DBPort:                  getEnv("DB_PORT", "3306"),
		DBUsername:              getEnv("DB_USERNAME", "mysql"),
		DBPassword:              getEnv("DB_PASSWORD", "mysql"),
		DBDatabase:              getEnv("DB_DATABASE", "ftoolbox"),
		Port:                    getEnv("PORT", "3000"),
		LogLevel:                getEnv("LOG_LEVEL", "info"),
		WorkerEnabled:           getEnvBool("WORKER_ENABLED", true),
		WorkerUpdateInterval:    getEnvInt("WORKER_UPDATE_INTERVAL", 10000),    // 10 seconds
		WorkerDiscoveryInterval: getEnvInt("WORKER_DISCOVERY_INTERVAL", 60000), // 60 seconds
		FanslyAPIRateLimit:      getEnvInt("FANSLY_API_RATE_LIMIT", 300),       // 60 requests per minute
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
