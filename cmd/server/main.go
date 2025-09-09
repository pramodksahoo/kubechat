package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pramodksahoo/kubechat/internal/config"
	"github.com/pramodksahoo/kubechat/internal/handlers"
	"github.com/pramodksahoo/kubechat/internal/k8s"
	"github.com/pramodksahoo/kubechat/internal/llm"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize Kubernetes client
	k8sClient, err := k8s.NewClient(cfg.KubeConfig)
	if err != nil {
		log.Fatal("Failed to initialize Kubernetes client:", err)
	}

	// Initialize LLM service
	llmService, err := llm.NewService(cfg.LLM)
	if err != nil {
		log.Fatal("Failed to initialize LLM service:", err)
	}

	// Set up Gin router
	r := gin.Default()

	// CORS middleware for development
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Serve static files (for React frontend)
	r.Static("/static", "./web/build/static")
	r.StaticFile("/", "./web/build/index.html")
	r.StaticFile("/favicon.ico", "./web/build/favicon.ico")

	// Initialize handlers
	h := handlers.NewHandlers(k8sClient, llmService)

	// API routes
	api := r.Group("/api")
	{
		api.POST("/query", h.HandleQuery)
		api.GET("/health", h.HandleHealth)
		api.GET("/clusters", h.HandleGetClusters)
		api.GET("/audit", h.HandleAuditLog)
		api.GET("/ws", h.HandleWebSocket)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting KubeChat server on port %s", port)
	log.Printf("LLM Provider: %s", cfg.LLM.Provider)
	
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}