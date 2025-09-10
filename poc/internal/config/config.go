package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	KubeConfig string
	LLM        LLMConfig
	Database   DatabaseConfig
	Server     ServerConfig
}

type LLMConfig struct {
	Provider string // "ollama" or "openai"
	
	// Ollama configuration
	OllamaURL   string
	OllamaModel string
	
	// OpenAI configuration  
	OpenAIKey   string
	OpenAIModel string
	
	// Fallback settings
	EnableFallback bool
}

type DatabaseConfig struct {
	Type string // "memory" for PoC, "postgres" for production
	URL  string
}

type ServerConfig struct {
	Port string
	Host string
}

func Load() *Config {
	// Default kubeconfig path
	kubeConfig := os.Getenv("KUBECONFIG")
	if kubeConfig == "" {
		home, _ := os.UserHomeDir()
		kubeConfig = filepath.Join(home, ".kube", "config")
	}

	return &Config{
		KubeConfig: kubeConfig,
		LLM: LLMConfig{
			Provider:       getEnv("LLM_PROVIDER", "ollama"),
			OllamaURL:      getEnv("OLLAMA_URL", "http://localhost:11434"),
			OllamaModel:    getEnv("OLLAMA_MODEL", "llama2"),
			OpenAIKey:      getEnv("OPENAI_API_KEY", ""),
			OpenAIModel:    getEnv("OPENAI_MODEL", "gpt-4"),
			EnableFallback: getEnv("LLM_FALLBACK", "true") == "true",
		},
		Database: DatabaseConfig{
			Type: getEnv("DB_TYPE", "memory"),
			URL:  getEnv("DB_URL", ""),
		},
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Host: getEnv("HOST", "0.0.0.0"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}