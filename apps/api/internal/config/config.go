package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for our application
type Config struct {
	Environment  string
	Port         int
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int

	// Database configuration
	DatabaseURL      string
	DatabaseName     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseHost     string
	DatabasePort     int
	DatabaseSSLMode  string

	// Redis configuration
	RedisURL      string
	RedisPassword string
	RedisDB       int

	// Kubernetes configuration
	KubeConfigPath string
	InCluster      bool

	// AI service configuration
	OllamaURL    string
	OpenAIAPIKey string

	// Logging configuration
	LogLevel  string
	LogFormat string
}

// Load returns a new Config struct with values from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Environment:  getEnv("ENVIRONMENT", "development"),
		Port:         getEnvAsInt("PORT", 8080),
		ReadTimeout:  getEnvAsInt("READ_TIMEOUT", 30),
		WriteTimeout: getEnvAsInt("WRITE_TIMEOUT", 30),
		IdleTimeout:  getEnvAsInt("IDLE_TIMEOUT", 120),

		// Database
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		DatabaseName:     getEnv("DATABASE_NAME", "kubechat"),
		DatabaseUser:     getEnv("DATABASE_USER", "postgres"),
		DatabasePassword: getEnv("DATABASE_PASSWORD", ""),
		DatabaseHost:     getEnv("DATABASE_HOST", "localhost"),
		DatabasePort:     getEnvAsInt("DATABASE_PORT", 5432),
		DatabaseSSLMode:  getEnv("DATABASE_SSL_MODE", "disable"),

		// Redis
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		// Kubernetes
		KubeConfigPath: getEnv("KUBE_CONFIG_PATH", ""),
		InCluster:      getEnvAsBool("IN_CLUSTER", false),

		// AI Services
		OllamaURL:    getEnv("OLLAMA_URL", "http://localhost:11434"),
		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),

		// Logging
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),
	}

	return cfg, nil
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}