# Environment Configuration Loading Pattern

## When to Use
Use this pattern for loading and managing application configuration from environment variables with sensible defaults.

## Why It Exists
Provides centralized configuration management with environment variable support, default values, and type-safe configuration loading. Ensures consistent configuration across all application components.

## Implementation

### Configuration Structure
Define all configuration fields in a single struct:
```go
type Config struct {
    DBHost                   string
    DBPort                   string
    DBUsername               string
    DBPassword               string
    DBDatabase               string
    Port                     string
    LogLevel                 string
    WorkerEnabled            bool
    WorkerUpdateInterval     int
    WorkerDiscoveryInterval  int
    RankCalculationInterval  int
    WorkerStatisticsInterval int
    GlobalRateLimit          int
    GlobalRateLimitWindow    int
}
```

### Loading Function
Load environment variables with defaults:
```go
func Load() *Config {
    godotenv.Load()
    
    return &Config{
        DBHost:                   getEnv("DB_HOST", "localhost"),
        DBPort:                   getEnv("DB_PORT", "3306"),
        DBUsername:               getEnv("DB_USERNAME", "mysql"),
        DBPassword:               getEnv("DB_PASSWORD", "mysql"),
        DBDatabase:               getEnv("DB_DATABASE", "ftoolbox"),
        Port:                     getEnv("PORT", "3000"),
        LogLevel:                 getEnv("LOG_LEVEL", "info"),
        WorkerEnabled:            getEnvBool("WORKER_ENABLED", true),
        WorkerUpdateInterval:     getEnvInt("WORKER_UPDATE_INTERVAL", 10000),
        WorkerDiscoveryInterval:  getEnvInt("WORKER_DISCOVERY_INTERVAL", 60000),
        RankCalculationInterval:  getEnvInt("RANK_CALCULATION_INTERVAL", 60000),
        WorkerStatisticsInterval: getEnvInt("WORKER_STATISTICS_INTERVAL", 3600000),
        GlobalRateLimit:          getEnvInt("FANSLY_GLOBAL_RATE_LIMIT", 50),
        GlobalRateLimitWindow:    getEnvInt("FANSLY_GLOBAL_RATE_LIMIT_WINDOW", 10),
    }
}
```

### Helper Functions
Type-safe environment variable parsing:
```go
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
```

## Source References
- `backend-go/config/config.go:10-46` - Configuration struct and loading function
- `backend-go/main.go:26` - Configuration loading in main function
- `backend-go/config/config.go:48-67` - Helper functions for type-safe parsing

## Key Conventions
- Always call `godotenv.Load()` first to load .env files
- Provide sensible defaults for all configuration values
- Use descriptive environment variable names in UPPER_SNAKE_CASE
- Implement type-safe parsing with fallback to defaults
- Load configuration once at application startup
- Pass configuration explicitly to components that need it
- Use milliseconds for time-based intervals to maintain consistency