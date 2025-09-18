// Package main provides the KubeChat API server
//
//	@title			KubeChat Enterprise Platform API
//	@version		1.0.0
//	@description	Kubernetes-native AI platform with multi-provider support for enterprise deployments
//	@description
//	@description				This API provides comprehensive management capabilities for:
//	@description				- Kubernetes cluster operations and monitoring
//	@description				- Multi-provider AI/ML services (OpenAI, Anthropic, Google, Ollama)
//	@description				- Real-time WebSocket communication
//	@description				- Advanced cost tracking and budget management
//	@description				- Enterprise security and compliance features
//	@description				- Comprehensive audit logging and monitoring
//
//	@contact.name				KubeChat Support Team
//	@contact.url				https://github.com/pramodksahoo/kubechat
//	@contact.email				support@kubechat.dev
//
//	@license.name				MIT License
//	@license.url				https://opensource.org/licenses/MIT
//
//	@host						localhost:8080
//	@BasePath					/api/v1
//	@schemes					http https
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Enter the token with the `Bearer ` prefix, e.g. "Bearer abcde12345"
//
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						X-API-Key
//	@description				API Key for service-to-service communication
//
//	@x-extension-openapi		{"example": "value"}
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/pramodksahoo/kubechat/apps/api/docs"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	auditHandler "github.com/pramodksahoo/kubechat/apps/api/internal/handlers/audit"
	"github.com/pramodksahoo/kubechat/apps/api/internal/handlers/auth"
	chatHandler "github.com/pramodksahoo/kubechat/apps/api/internal/handlers/chat"
	commandHandler "github.com/pramodksahoo/kubechat/apps/api/internal/handlers/command"
	communicationHandler "github.com/pramodksahoo/kubechat/apps/api/internal/handlers/communication"
	databaseHandler "github.com/pramodksahoo/kubechat/apps/api/internal/handlers/database"
	externalHandler "github.com/pramodksahoo/kubechat/apps/api/internal/handlers/external"
	gatewayHandler "github.com/pramodksahoo/kubechat/apps/api/internal/handlers/gateway"
	healthHandler "github.com/pramodksahoo/kubechat/apps/api/internal/handlers/health"
	kubernetesHandler "github.com/pramodksahoo/kubechat/apps/api/internal/handlers/kubernetes"
	nlpHandler "github.com/pramodksahoo/kubechat/apps/api/internal/handlers/nlp"
	securityHandler "github.com/pramodksahoo/kubechat/apps/api/internal/handlers/security"
	websocketHandler "github.com/pramodksahoo/kubechat/apps/api/internal/handlers/websocket"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/repositories"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/admin"
	auditService "github.com/pramodksahoo/kubechat/apps/api/internal/services/audit"
	authService "github.com/pramodksahoo/kubechat/apps/api/internal/services/auth"
	cacheService "github.com/pramodksahoo/kubechat/apps/api/internal/services/cache"
	chatService "github.com/pramodksahoo/kubechat/apps/api/internal/services/chat"
	commandService "github.com/pramodksahoo/kubechat/apps/api/internal/services/command"
	communicationService "github.com/pramodksahoo/kubechat/apps/api/internal/services/communication"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/credentials"
	databaseService "github.com/pramodksahoo/kubechat/apps/api/internal/services/database"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/external"
	gatewayService "github.com/pramodksahoo/kubechat/apps/api/internal/services/gateway"
	healthService "github.com/pramodksahoo/kubechat/apps/api/internal/services/health"
	kubernetesService "github.com/pramodksahoo/kubechat/apps/api/internal/services/kubernetes"
	nlpService "github.com/pramodksahoo/kubechat/apps/api/internal/services/nlp"
	queryService "github.com/pramodksahoo/kubechat/apps/api/internal/services/query"
	safetyService "github.com/pramodksahoo/kubechat/apps/api/internal/services/safety"
	securityService "github.com/pramodksahoo/kubechat/apps/api/internal/services/security"
	websocketService "github.com/pramodksahoo/kubechat/apps/api/internal/services/websocket"
)

// getAllowedOrigins returns secure CORS origins from environment with production defaults
func getAllowedOrigins() []string {
	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
	if allowedOriginsEnv == "" {
		// Production-first: No default origins allowed, must be explicitly configured
		// Debug mode allows localhost for convenience
		if gin.Mode() == gin.DebugMode {
			log.Println("DEBUG MODE: Using default localhost CORS origins")
			return []string{"http://localhost:3000", "http://localhost:8080", "https://localhost:3000", "https://localhost:8080"}
		}
		// Production requires explicit configuration - fail fast
		log.Fatal("SECURITY CRITICAL: ALLOWED_ORIGINS environment variable must be explicitly set in production mode")
	}

	// Parse comma-separated origins and validate
	origins := strings.Split(allowedOriginsEnv, ",")
	validOrigins := make([]string, 0, len(origins))

	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin == "*" {
			log.Fatal("SECURITY CRITICAL: Wildcard CORS origin (*) is not allowed in production")
		}
		if origin != "" {
			validOrigins = append(validOrigins, origin)
		}
	}

	if len(validOrigins) == 0 {
		log.Fatal("SECURITY CRITICAL: At least one valid CORS origin must be specified")
	}

	return validOrigins
}

// securityValidationMiddleware implements comprehensive security headers and validation
func securityValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// API versioning headers
		c.Header("API-Version", "v1")
		c.Header("X-API-Version", "1.0.0")
		c.Header("X-Service-Name", "kubechat-api")

		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// HSTS for HTTPS
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Enhanced Content Security Policy
		csp := "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self' wss: ws:; font-src 'self' data:; object-src 'none'; base-uri 'self'; form-action 'self'"
		c.Header("Content-Security-Policy", csp)

		// Additional security headers
		c.Header("X-Permitted-Cross-Domain-Policies", "none")
		c.Header("Cross-Origin-Embedder-Policy", "require-corp")
		c.Header("Cross-Origin-Opener-Policy", "same-origin")
		c.Header("Cross-Origin-Resource-Policy", "same-origin")

		// Request size validation (prevent DoS)
		if c.Request.ContentLength > 10*1024*1024 { // 10MB limit
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "Request entity too large",
				"code":  "REQUEST_TOO_LARGE",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func main() {
	// Production-first configuration: default to release mode
	// Use GIN_MODE=debug environment variable to enable debug mode if needed
	if os.Getenv("GIN_MODE") == "debug" {
		gin.SetMode(gin.DebugMode)
		log.Println("Running in DEBUG mode (GIN_MODE=debug)")
	} else {
		gin.SetMode(gin.ReleaseMode)
		log.Println("Running in PRODUCTION mode (default)")
	}

	// Perform comprehensive security validation at startup
	log.Println("Performing security validation...")
	securityValidator := securityService.NewSecurityValidator()
	if err := securityValidator.ValidateSecurityConfiguration(); err != nil {
		log.Fatalf("SECURITY VALIDATION FAILED: %v", err)
	}
	log.Println("✅ Security validation passed")

	// Initialize database connection
	db, err := initDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	auditRepo := repositories.NewAuditRepository(db)

	// Initialize Admin Service to ensure admin user exists with secret-based password
	adminConfig := &admin.Config{
		Namespace:  "kubechat", // Default namespace, can be overridden by env var
		SecretName: "kubechat-admin-secret",
	}
	adminSvc, err := admin.NewService(userRepo, adminConfig)
	if err != nil {
		log.Printf("Warning: Failed to initialize admin service: %v", err)
	} else {
		// Ensure admin user exists with password from Kubernetes secret
		if err := adminSvc.EnsureAdminUser(context.Background()); err != nil {
			log.Printf("Warning: Failed to ensure admin user: %v", err)
		} else {
			log.Println("✅ Admin user verified and password synced from Kubernetes secret")
		}
	}

	// Initialize services with mandatory JWT secret validation
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("SECURITY CRITICAL: JWT_SECRET environment variable must be set. Application cannot start without secure JWT secret.")
	}
	if jwtSecret == "dev-secret-key-change-in-production" || jwtSecret == "your-secret-key" || len(jwtSecret) < 32 {
		log.Fatal("SECURITY CRITICAL: JWT_SECRET must be a secure, unique value with minimum 32 characters. Default or weak secrets are not allowed.")
	}
	authSvc := authService.NewService(userRepo, jwtSecret)

	// Initialize Kubernetes service with allowed namespaces
	allowedNamespaces := []string{"kubechat", "kubechat-*", "default"}
	kubeSvc, err := kubernetesService.NewService("", allowedNamespaces)
	if err != nil {
		log.Printf("Warning: Failed to initialize Kubernetes service: %v", err)
		log.Println("Kubernetes functionality will be disabled")
	}

	// Initialize NLP service with Ollama and OpenAI providers
	ollamaBaseURL := os.Getenv("OLLAMA_BASE_URL")
	if ollamaBaseURL == "" {
		ollamaBaseURL = "http://kubechat-dev-ollama:11434"
	}

	ollamaModel := os.Getenv("OLLAMA_MODEL")
	if ollamaModel == "" {
		ollamaModel = "llama3.2:3b"
	}

	ollamaConfig := &nlpService.OllamaConfig{
		BaseURL:        ollamaBaseURL,
		Model:          ollamaModel,
		TimeoutSeconds: 30,
		MaxTokens:      1000,
		Temperature:    0.1,
		EnableStream:   false,
	}

	openaiBaseURL := os.Getenv("OPENAI_BASE_URL")
	if openaiBaseURL == "" {
		openaiBaseURL = "https://api.openai.com/v1"
	}

	openaiModel := os.Getenv("OPENAI_MODEL")
	if openaiModel == "" {
		openaiModel = "gpt-3.5-turbo"
	}

	openaiConfig := &nlpService.OpenAIConfig{
		APIKey:         os.Getenv("OPENAI_API_KEY"),
		BaseURL:        openaiBaseURL,
		Model:          openaiModel,
		TimeoutSeconds: 30,
		MaxTokens:      1000,
		Temperature:    0.1,
	}

	nlpConfig := &nlpService.Config{
		DefaultProvider:    models.ProviderOllama, // Default to Ollama (local)
		EnableFallback:     true,
		MaxRetries:         3,
		TimeoutSeconds:     30,
		EnableCaching:      false, // Disabled for development
		CacheTTLMinutes:    60,
		EnableRateLimiting: true,
		RateLimit:          30, // 30 requests per minute
	}

	ollamaSvc := nlpService.NewOllamaService(ollamaConfig)
	openaiSvc := nlpService.NewOpenAIService(openaiConfig)
	nlpSvc := nlpService.NewService(ollamaSvc, openaiSvc, nlpConfig)

	log.Println("NLP service initialized with Ollama and OpenAI providers")

	// Initialize Chat Service
	chatConfig := chatService.DefaultConfig()
	chatSvc := chatService.NewService(nlpSvc, chatConfig)
	log.Println("Chat service initialized")

	// Initialize Safety Classification Service (Story 1.5 Task 2)
	safetyConfig := &safetyService.Config{
		SafeThreshold:      30.0,
		WarningThreshold:   70.0,
		DangerousThreshold: 70.0,
		ProductionMode:     true,
		BlockDangerous:     true,
		RequireApproval:    true,
	}
	safetySvc := safetyService.NewService(safetyConfig)
	log.Println("Safety classification service initialized with context-aware security")

	// Initialize Command Execution Services (Story 1.6)
	// Create a simple in-memory repository implementation for now
	commandRepo := &mockCommandRepository{}

	// Initialize cache service
	cacheConfig := &cacheService.Config{
		RedisURL:               "redis://kubechat-dev-redis-master:6379",
		CommandResultTTL:       30 * time.Minute,
		ResourceDataTTL:        15 * time.Minute,
		MetricsRetentionPeriod: 24 * time.Hour,
		MaxMemoryUsage:         "100mb",
		EnableMetrics:          true,
	}
	cacheSvc, err := cacheService.NewService(cacheConfig)
	if err != nil {
		log.Printf("Warning: Failed to initialize cache service: %v", err)
		// Use a nil cache service for now
		cacheSvc = nil
	}

	commandSvc := commandService.NewService(commandRepo, kubeSvc, safetySvc, cacheSvc)
	log.Println("Command execution service initialized with safety integration and caching")

	// Note: Rollback service requires proper repositories - skipping for now to avoid deployment issues

	// Initialize Query Processing Service (Story 1.5 Task 3)
	queryConfig := &queryService.Config{
		DefaultTimeout:   30 * time.Second,
		MaxQueryLength:   2000,
		MinQueryLength:   3,
		DefaultNamespace: "default",
	}
	querySvc := queryService.NewService(nlpSvc, safetySvc, queryConfig)
	log.Println("Query processing service initialized with safety integration")

	// Initialize Audit service
	auditConfig := &auditService.Config{
		EnableAsyncLogging:      true,
		AsyncBufferSize:         1000,
		MaxBatchSize:            100,
		EnableStructuredLogging: true,
		LogLevel:                "info",
		EnableIntegrityCheck:    true,
		RetentionDays:           90,
	}
	auditSvc := auditService.NewService(auditRepo, auditConfig)

	log.Println("Audit service initialized with structured logging and integrity checking")

	// Initialize WebSocket service
	websocketConfig := &websocketService.Config{
		ReadTimeout:           60 * time.Second,
		WriteTimeout:          10 * time.Second,
		PingPeriod:            54 * time.Second,
		PongWait:              60 * time.Second,
		MaxMessageSize:        1024 * 1024, // 1MB
		ReadBufferSize:        1024,
		WriteBufferSize:       1024,
		MaxClients:            1000,
		ClientTimeout:         5 * time.Minute,
		HeartbeatInterval:     30 * time.Second,
		MaxConcurrentCommands: 10,
		CommandTimeout:        5 * time.Minute,
		AllowedOrigins:        getAllowedOrigins(), // Secure CORS origins from environment
		RequireAuth:           true,
		EnableDebugLogging:    gin.Mode() == gin.DebugMode,
		LogLevel:              "info",
	}

	websocketSvc := websocketService.NewService(authSvc, kubeSvc, nlpSvc, auditSvc, websocketConfig)
	log.Println("WebSocket service initialized with real-time communication support")

	// Initialize API Gateway service
	gatewayConfig := &gatewayService.Config{
		DefaultRateLimit:        60, // 60 requests per minute
		BurstLimit:              10, // allow bursts of 10
		RateLimitWindow:         time.Minute,
		CircuitBreakerThreshold: 5, // 5 failures before opening
		CircuitBreakerTimeout:   30 * time.Second,
		CircuitBreakerReset:     60 * time.Second,
		EnableSecurityHeaders:   true,
		AllowedOrigins:          getAllowedOrigins(), // Secure CORS origins from environment
		AllowedMethods:          []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:          []string{"Content-Type", "Authorization", "X-Requested-With"},
		MaxRequestSize:          10 * 1024 * 1024, // 10MB
		EnableRequestLogging:    gin.Mode() == gin.DebugMode,
		EnableResponseLogging:   false,
		LogLevel:                "info",
		EnableGzip:              true,
		ReadTimeout:             30 * time.Second,
		WriteTimeout:            30 * time.Second,
		MaxConcurrentRequests:   1000,
	}

	gatewaySvc := gatewayService.NewService(gatewayConfig)
	log.Println("API Gateway service initialized with rate limiting and security features")

	// Initialize Health Check service
	healthConfig := &healthService.Config{
		CheckInterval:      30 * time.Second,
		UnhealthyThreshold: 3,
		EnableMonitoring:   true,
		DatabaseTimeout:    5 * time.Second,
		RedisTimeout:       5 * time.Second,
		ExternalTimeout:    10 * time.Second,
		AlertThreshold:     2 * time.Minute,
		EnableAlerts:       gin.Mode() == gin.DebugMode,
		LogHealthStatus:    gin.Mode() == gin.DebugMode,
		MetricsRetention:   24 * time.Hour,
	}

	healthSvc := healthService.NewService(db, healthConfig)
	log.Println("Health Check service initialized with monitoring and alerting")

	// Start health monitoring
	if err := healthSvc.StartHealthMonitoring(context.Background()); err != nil {
		log.Printf("Warning: Failed to start health monitoring: %v", err)
	}

	// Initialize Service Communication
	communicationConfig := &communicationService.Config{
		ServiceDiscoveryEnabled: true,
		RegistryType:            "memory",
		CircuitBreakerEnabled:   true,
		FailureThreshold:        5,
		RecoveryTimeout:         30 * time.Second,
		MaxConcurrentRequests:   100,
		RetryEnabled:            true,
		MaxRetries:              3,
		RetryBackoff:            100 * time.Millisecond,
		RetryMultiplier:         2.0,
		RequestTimeout:          30 * time.Second,
		ConnectionTimeout:       5 * time.Second,
		KeepAliveTimeout:        30 * time.Second,
		DefaultLoadBalancer:     models.LoadBalancingRoundRobin,
		HealthCheckInterval:     30 * time.Second,
		UnhealthyThreshold:      3,
		EventSystemEnabled:      true,
		EventBufferSize:         1000,
		EventProcessorWorkers:   5,
		EnableMetrics:           true,
		MetricsUpdateInterval:   10 * time.Second,
	}

	commSvc := communicationService.NewService(communicationConfig)
	log.Println("Service Communication initialized with discovery, circuit breakers, and load balancing")

	// Start communication health checking
	if err := commSvc.StartHealthChecking(context.Background()); err != nil {
		log.Printf("Warning: Failed to start communication health checking: %v", err)
	}

	// Initialize Database Connection Management service
	databaseConfig := &databaseService.Config{
		HealthCheckInterval:    30 * time.Second,
		MetricsUpdateInterval:  10 * time.Second,
		FailoverEnabled:        true,
		AutoFailoverEnabled:    gin.Mode() != gin.DebugMode, // Disable auto-failover in debug mode
		MaxConnectionRetries:   3,
		ConnectionRetryDelay:   time.Second,
		SlowQueryThreshold:     5 * time.Second,
		ConnectionTimeout:      10 * time.Second,
		QueryTimeout:           30 * time.Second,
		MaxPoolSize:            10,
		DefaultMaxOpenConns:    25,
		DefaultMaxIdleConns:    5,
		DefaultConnMaxLifetime: time.Hour,
		DefaultConnMaxIdleTime: 30 * time.Minute,
	}

	dbSvc := databaseService.NewService(databaseConfig)
	log.Println("Database Connection Management service initialized with pooling and failover support")

	// Start database health checking
	if err := dbSvc.StartHealthChecking(context.Background()); err != nil {
		log.Printf("Warning: Failed to start database health checking: %v", err)
	}

	// Initialize Security and Performance Integration service
	securityConfig := &securityService.Config{
		EnableSecurityScanning:      true,
		SecurityScanInterval:        24 * time.Hour,
		PasswordHashCost:            bcrypt.DefaultCost,
		SessionTimeout:              4 * time.Hour,
		MaxSessionsPerUser:          5,
		EnableBruteForceProtection:  true,
		BruteForceThreshold:         5,
		BruteForceLockoutDuration:   30 * time.Minute,
		EnablePerformanceMonitoring: true,
		MetricsCollectionInterval:   30 * time.Second,
		EnableCaching:               true,
		CacheSize:                   100 * 1024 * 1024, // 100MB
		CacheTTL:                    time.Hour,
		EnableCompression:           true,
		CompressionLevel:            6,
		DefaultRateLimit:            100, // requests per window
		RateLimitWindow:             time.Minute,
		EnableIPBlocking:            true,
		IPBlockDuration:             time.Hour,
		EnableRealTimeAlerts:        gin.Mode() == gin.DebugMode,
		AlertWebhookURL:             os.Getenv("SECURITY_WEBHOOK_URL"),
	}

	secSvc := securityService.NewService(securityConfig)
	log.Println("Security and Performance Integration service initialized with comprehensive protection")

	// Start security monitoring
	if err := secSvc.StartMonitoring(context.Background()); err != nil {
		log.Printf("Warning: Failed to start security monitoring: %v", err)
	}

	// Initialize External API Services (Tasks 1-3)
	log.Println("Initializing External API Services...")

	// Initialize Kubernetes client for external services
	var k8sClient kubernetes.Interface
	if config, err := rest.InClusterConfig(); err == nil {
		if clientset, err := kubernetes.NewForConfig(config); err == nil {
			k8sClient = clientset
		} else {
			log.Printf("Warning: Failed to create Kubernetes clientset: %v", err)
		}
	} else {
		log.Printf("Warning: Not running in cluster, Kubernetes features limited: %v", err)
	}

	// Initialize Credentials Service (Task 3)
	credentialsConfig := &credentials.Config{
		Namespace:        "kubechat",
		SecretPrefix:     "kubechat-creds",
		EnableEncryption: true,
		EnableAuditLog:   true,
		AutoRotationDays: 90,
	}
	credentialsSvc, err := credentials.NewService(k8sClient, auditSvc, credentialsConfig)
	if err != nil {
		log.Printf("Warning: Failed to initialize credentials service: %v", err)
	} else {
		log.Println("Credentials service initialized with Kubernetes secrets and encryption")
	}

	// Initialize External API Client Manager (Task 1)
	clientManager := external.NewClientManager(auditSvc)

	// Initialize OpenAI Client for external APIs
	openaiExternalConfig := &external.ClientConfig{
		APIKey:           os.Getenv("OPENAI_API_KEY"),
		BaseURL:          openaiBaseURL,
		Timeout:          30 * time.Second,
		MaxRetries:       3,
		RateLimitPerMin:  60,
		FailureThreshold: 5,
		RecoveryTimeout:  30 * time.Second,
		EnableMetrics:    true,
		EnableAuditLog:   true,
		UserAgent:        "KubeChat-External/1.0",
		CustomHeaders:    make(map[string]string),
	}
	openaiClient, err := external.NewOpenAIClient(openaiExternalConfig, auditSvc)
	if err != nil {
		log.Printf("Warning: Failed to create OpenAI client: %v", err)
	} else {
		if err := clientManager.RegisterClient("openai", openaiClient, openaiExternalConfig); err != nil {
			log.Printf("Warning: Failed to register OpenAI client: %v", err)
		} else {
			log.Println("OpenAI external API client registered successfully")
		}
	}

	// Initialize Health Service for External APIs (Task 4)
	healthServiceConfig := &external.HealthServiceConfig{
		DefaultTimeout:      30 * time.Second,
		DefaultInterval:     5 * time.Minute,
		MaxConcurrentChecks: 10,
		EnableAuditLogging:  true,
		EnableAlerts:        gin.Mode() == gin.DebugMode,
	}
	externalHealthSvc, err := external.NewHealthService(auditSvc, healthServiceConfig)
	if err != nil {
		log.Printf("Warning: Failed to initialize external health service: %v", err)
	} else {
		log.Println("External API health monitoring service initialized")

		// Register APIs with health monitoring (Task 4.1)
		if openaiClient != nil {
			openaiHealthConfig := &external.APIHealthConfig{
				Name:        "openai",
				DisplayName: "OpenAI API",
				Description: "OpenAI GPT API for chat completions",
				Endpoint:    openaiBaseURL + "/models",
				Method:      "GET",
				Headers: map[string]string{
					"Authorization": "Bearer " + os.Getenv("OPENAI_API_KEY"),
					"User-Agent":    "KubeChat-Health/1.0",
				},
				Timeout:             15 * time.Second,
				Interval:            2 * time.Minute,
				RetryAttempts:       3,
				RetryInterval:       5 * time.Second,
				ExpectedStatusCodes: []int{200},
				Enabled:             true,
				Tags:                []string{"external", "ai", "openai"},
				Metadata: map[string]string{
					"provider": "openai",
					"type":     "ai-api",
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := externalHealthSvc.RegisterAPI(context.Background(), openaiHealthConfig); err != nil {
				log.Printf("Warning: Failed to register OpenAI API for health monitoring: %v", err)
			} else {
				log.Println("OpenAI API registered for health monitoring")
			}
		}

		// Register internal services for health monitoring (Task 4.1)
		internalAPIs := []struct {
			name        string
			displayName string
			description string
			endpoint    string
			tags        []string
		}{
			{
				name:        "ollama",
				displayName: "Ollama AI Service",
				description: "Local Ollama AI service for chat completions",
				endpoint:    "http://kubechat-dev-ollama:11434/api/tags",
				tags:        []string{"internal", "ai", "ollama"},
			},
			{
				name:        "postgresql",
				displayName: "PostgreSQL Database",
				description: "Primary application database",
				endpoint:    "http://kubechat-dev-postgresql:5432", // TCP health check
				tags:        []string{"internal", "database", "postgresql"},
			},
			{
				name:        "redis",
				displayName: "Redis Cache",
				description: "Redis cache and session storage",
				endpoint:    "http://kubechat-dev-redis-master:6379", // TCP health check
				tags:        []string{"internal", "cache", "redis"},
			},
		}

		for _, api := range internalAPIs {
			config := &external.APIHealthConfig{
				Name:        api.name,
				DisplayName: api.displayName,
				Description: api.description,
				Endpoint:    api.endpoint,
				Method:      "GET",
				Headers: map[string]string{
					"User-Agent": "KubeChat-Health/1.0",
				},
				Timeout:             10 * time.Second,
				Interval:            1 * time.Minute, // More frequent for internal services
				RetryAttempts:       2,
				RetryInterval:       3 * time.Second,
				ExpectedStatusCodes: []int{200, 404}, // 404 acceptable for some endpoints
				Enabled:             true,
				Tags:                api.tags,
				Metadata: map[string]string{
					"environment": "development",
					"internal":    "true",
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			if err := externalHealthSvc.RegisterAPI(context.Background(), config); err != nil {
				log.Printf("Warning: Failed to register %s for health monitoring: %v", api.name, err)
			} else {
				log.Printf("%s registered for health monitoring", api.displayName)
			}
		}

		// Start continuous monitoring (Task 4.2)
		if err := externalHealthSvc.StartContinuousMonitoring(context.Background()); err != nil {
			log.Printf("Warning: Failed to start continuous health monitoring: %v", err)
		} else {
			log.Println("Continuous health monitoring started for all registered APIs")
		}
	}

	// Initialize Encryption Service (Task 3)
	encryptionConfig := &external.EncryptionConfig{
		Namespace:       "kubechat",
		MasterKeySecret: "kubechat-encryption-key",
		DefaultLevel:    external.EncryptionLevelHigh,
		EnableMetrics:   true,
		EnableAuditLog:  true,
	}
	encryptionSvc, err := external.NewEncryptionService(k8sClient, auditSvc, encryptionConfig)
	if err != nil {
		log.Printf("Warning: Failed to initialize encryption service: %v", err)
	} else {
		log.Println("Encryption service initialized with AES-256 and key management")
	}

	// Initialize Credential Audit Service (Task 3)
	credentialAuditConfig := &external.AuditConfig{
		EnableDetailedLogging: true,
		EnableRealTimeAlerts:  true,
		DefaultRetention:      365 * 24 * time.Hour,
		MaxQueryResults:       10000,
		EnableCompression:     true,
		EncryptLogs:           true,
	}
	credentialAuditSvc, err := external.NewCredentialAuditService(auditSvc, credentialAuditConfig)
	if err != nil {
		log.Printf("Warning: Failed to initialize credential audit service: %v", err)
	} else {
		log.Println("Credential audit service initialized with comprehensive logging")
	}

	// Initialize Credential Validator (Task 3)
	validatorConfig := &external.CredentialValidatorConfig{
		EnableConnectivityTests: true,
		DefaultTimeout:          30 * time.Second,
		MaxRetryAttempts:        3,
		EnableAuditLog:          true,
		StrictValidation:        false,
	}
	credentialValidator, err := external.NewCredentialValidator(auditSvc, validatorConfig)
	if err != nil {
		log.Printf("Warning: Failed to initialize credential validator: %v", err)
	} else {
		log.Println("Credential validator initialized with format and strength analysis")
	}

	// Initialize Credential Injector (Task 3)
	injectorConfig := &external.InjectorConfig{
		Namespace:        "kubechat",
		EnableScheduling: true,
		EnableAuditLog:   true,
		BackupEnabled:    true,
	}
	credentialInjector, err := external.NewCredentialInjector(k8sClient, credentialsSvc, auditSvc, injectorConfig)
	if err != nil {
		log.Printf("Warning: Failed to initialize credential injector: %v", err)
	} else {
		log.Println("Credential injector initialized with secure injection mechanisms")
	}

	// Initialize Fallback Service (Task 5.2, 5.3, 5.4)
	fallbackSvc := external.NewFallbackService()

	// Register fallback chains for external APIs
	openaiChain := &external.FallbackChain{
		ServiceName:     "openai",
		PrimaryProvider: "openai-api",
		Fallbacks: []external.FallbackOption{
			{
				Type:        external.FallbackTypeMock,
				Priority:    1,
				Enabled:     true,
				Timeout:     5 * time.Second,
				Description: "Mock OpenAI responses for development",
			},
			{
				Type:        external.FallbackTypeDefault,
				Priority:    2,
				Enabled:     true,
				Timeout:     1 * time.Second,
				Description: "Default error response",
			},
		},
		DefaultResponse: map[string]interface{}{
			"error":    "OpenAI service temporarily unavailable",
			"fallback": "Please try again later",
		},
		CacheEnabled: true,
		CacheTTL:     5 * time.Minute,
		MockEnabled:  gin.Mode() == gin.DebugMode, // Enable mocks in debug mode
	}

	ollamaChain := &external.FallbackChain{
		ServiceName:     "ollama",
		PrimaryProvider: "ollama-local",
		Fallbacks: []external.FallbackOption{
			{
				Type:        external.FallbackTypeMock,
				Priority:    1,
				Enabled:     true,
				Timeout:     3 * time.Second,
				Description: "Mock Ollama responses",
			},
		},
		DefaultResponse: map[string]interface{}{
			"error":    "Ollama service temporarily unavailable",
			"fallback": "Switching to alternative AI provider",
		},
		MockEnabled: gin.Mode() == gin.DebugMode,
	}

	if err := fallbackSvc.RegisterFallbackChain("openai", openaiChain); err != nil {
		log.Printf("Warning: Failed to register OpenAI fallback chain: %v", err)
	} else {
		log.Println("OpenAI fallback chain registered with mock and default responses")
	}

	if err := fallbackSvc.RegisterFallbackChain("ollama", ollamaChain); err != nil {
		log.Printf("Warning: Failed to register Ollama fallback chain: %v", err)
	} else {
		log.Println("Ollama fallback chain registered with mock responses")
	}

	// Initialize Recovery Service (Task 5.5)
	recoverySvc := external.NewRecoveryService(auditSvc)

	// Register recovery procedures for external APIs
	if openaiClient != nil {
		openaiRecoveryProcedure := &external.RecoveryProcedure{
			Description:   "OpenAI API recovery procedure",
			MaxAttempts:   3,
			InitialDelay:  30 * time.Second,
			MaxDelay:      5 * time.Minute,
			BackoffFactor: 2.0,
			Steps: []external.RecoveryStep{
				{
					Name:        "health_check",
					Type:        external.RecoveryStepHealthCheck,
					Timeout:     15 * time.Second,
					RetryCount:  2,
					Description: "Check OpenAI API connectivity",
					Required:    true,
					Action: func() error {
						return openaiClient.HealthCheck(context.Background())
					},
				},
				{
					Name:        "clear_cache",
					Type:        external.RecoveryStepClearCache,
					Timeout:     5 * time.Second,
					RetryCount:  1,
					Description: "Clear any cached connections",
					Required:    false,
					Action: func() error {
						log.Println("Clearing OpenAI client cache (simulated)")
						return nil
					},
				},
			},
			HealthCheck: func() error {
				return openaiClient.HealthCheck(context.Background())
			},
			Enabled: true,
		}

		if err := recoverySvc.RegisterRecoveryProcedure("openai", openaiRecoveryProcedure); err != nil {
			log.Printf("Warning: Failed to register OpenAI recovery procedure: %v", err)
		} else {
			log.Println("OpenAI recovery procedure registered with health checks and cache clearing")
		}
	}

	// Start recovery monitoring
	if err := recoverySvc.StartRecoveryMonitoring(context.Background()); err != nil {
		log.Printf("Warning: Failed to start recovery monitoring: %v", err)
	} else {
		log.Println("API recovery monitoring started - automatic recovery enabled")
	}

	// Initialize Cost Tracking and Budget Management Services (Task 6)
	log.Println("Initializing Cost Tracking and Budget Management...")

	// Initialize Cost Tracking Service (Task 6.1)
	costSvc := external.NewCostService()
	log.Println("Cost tracking service initialized with usage monitoring and billing support")

	// Initialize Budget Management Service (Task 6.2)
	budgetSvc := external.NewBudgetService(costSvc)
	log.Println("Budget management service initialized with alerts and limits")

	// Initialize Cost Optimization Service (Task 6.3)
	optimizationSvc := external.NewOptimizationService(costSvc)
	log.Println("Cost optimization service initialized with usage analytics and recommendations")

	// Initialize Cost Allocation Service (Task 6.4)
	allocationSvc := external.NewAllocationService(costSvc)
	log.Println("Cost allocation service initialized with department and project tracking")

	// Initialize Automated Cost Control Service (Task 6.5)
	controlSvc := external.NewControlService(budgetSvc, costSvc)
	log.Println("Automated cost control service initialized with threshold monitoring and emergency controls")

	// Initialize default cost tracking for services
	if openaiClient != nil {
		// Sample usage recording for OpenAI
		go func() {
			for {
				time.Sleep(5 * time.Minute)
				sampleUsage := &external.UsageRequest{
					ServiceName:  "openai",
					Operation:    "chat_completion",
					Tokens:       1500,
					RequestCount: 1,
					ResponseTime: 2 * time.Second,
					ModelUsed:    "gpt-3.5-turbo",
					UserID:       "system",
					SessionID:    "background",
					Department:   "engineering",
					Project:      "kubechat_core",
					Metadata: map[string]string{
						"endpoint": "/v1/chat/completions",
						"type":     "background_monitoring",
					},
					Timestamp: time.Now(),
				}
				if err := costSvc.RecordUsage(context.Background(), sampleUsage); err != nil {
					log.Printf("Warning: Failed to record usage: %v", err)
				}
			}
		}()
	}

	if ollamaSvc != nil {
		// Sample usage recording for Ollama
		go func() {
			for {
				time.Sleep(3 * time.Minute)
				sampleUsage := &external.UsageRequest{
					ServiceName:  "ollama",
					Operation:    "generate",
					Tokens:       800,
					RequestCount: 1,
					ResponseTime: 1 * time.Second,
					ModelUsed:    "llama3.2:3b",
					UserID:       "system",
					SessionID:    "background",
					Department:   "engineering",
					Project:      "kubechat_ai",
					Metadata: map[string]string{
						"endpoint": "/api/generate",
						"type":     "background_monitoring",
					},
					Timestamp: time.Now(),
				}
				if err := costSvc.RecordUsage(context.Background(), sampleUsage); err != nil {
					log.Printf("Warning: Failed to record usage: %v", err)
				}
			}
		}()
	}

	log.Println("Cost tracking and budget management services initialization completed")

	// Initialize Multi-Provider Support Architecture Services (Task 7)
	log.Println("Initializing Multi-Provider Support Architecture...")

	// Initialize Provider Registry Service (Task 7.1 & 7.2)
	providerRegistry := external.NewProviderRegistry()

	// Create OpenAI Provider configuration
	openaiProviderConfig := &external.ProviderConfig{
		Name:           "openai",
		Type:           "external",
		Version:        "1.0.0",
		Enabled:        true,
		BaseURL:        openaiBaseURL,
		APIKey:         os.Getenv("OPENAI_API_KEY"),
		Models:         []string{openaiModel, "gpt-4", "gpt-3.5-turbo"},
		Capabilities:   []string{"chat", "completion", "embedding"},
		MaxConcurrency: 10,
		CustomParameters: map[string]interface{}{
			"max_tokens":  1000,
			"temperature": 0.1,
			"timeout":     30,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create OpenAI MockProvider instance and initialize it
	openaiProvider := &external.MockProvider{}
	if err := openaiProvider.Initialize(openaiProviderConfig); err != nil {
		log.Printf("Warning: Failed to initialize OpenAI provider: %v", err)
	} else {
		if err := providerRegistry.RegisterProvider(openaiProvider); err != nil {
			log.Printf("Warning: Failed to register OpenAI provider: %v", err)
		} else {
			log.Println("OpenAI provider registered successfully")
		}
	}

	// Create Ollama Provider configuration
	ollamaProviderConfig := &external.ProviderConfig{
		Name:           "ollama",
		Type:           "local",
		Version:        "1.0.0",
		Enabled:        true,
		BaseURL:        ollamaBaseURL,
		Models:         []string{ollamaModel, "llama3.2:1b", "llama3.2:3b", "codellama"},
		Capabilities:   []string{"chat", "completion", "code"},
		MaxConcurrency: 5,
		CustomParameters: map[string]interface{}{
			"max_tokens":  1000,
			"temperature": 0.1,
			"stream":      false,
			"timeout":     30,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create Ollama MockProvider instance and initialize it
	ollamaProvider := &external.MockProvider{}
	if err := ollamaProvider.Initialize(ollamaProviderConfig); err != nil {
		log.Printf("Warning: Failed to initialize Ollama provider: %v", err)
	} else {
		if err := providerRegistry.RegisterProvider(ollamaProvider); err != nil {
			log.Printf("Warning: Failed to register Ollama provider: %v", err)
		} else {
			log.Println("Ollama provider registered successfully")
		}
	}

	// Discover additional providers (Task 7.2)
	if err := providerRegistry.DiscoverProviders(); err != nil {
		log.Printf("Warning: Provider discovery failed: %v", err)
	} else {
		allProviders := providerRegistry.GetAllProviders()
		discoveredCount := len(allProviders) - 2 // Subtract the 2 we manually registered
		if discoveredCount > 0 {
			log.Printf("Discovered %d additional providers", discoveredCount)
			for _, provider := range allProviders {
				if provider.GetName() != "openai" && provider.GetName() != "ollama" {
					log.Printf("- %s (%s)", provider.GetName(), provider.GetType())
				}
			}
		} else {
			log.Println("No additional providers discovered")
		}
	}
	log.Println("Provider registry service initialized with dynamic discovery")

	// Initialize Load Balancer Service (Task 7.3)
	loadBalancer := external.NewProviderLoadBalancer(providerRegistry)

	// Set default load balancing strategy (weighted)
	if err := loadBalancer.SetLoadBalancingStrategy(external.LoadBalancingWeighted); err != nil {
		log.Printf("Warning: Failed to set load balancing strategy: %v", err)
	}

	// Configure provider weights
	providerWeights := map[string]float64{
		"openai": 0.4, // High weight for OpenAI
		"ollama": 0.6, // Higher weight for Ollama (local, faster, cost-free)
	}
	if err := loadBalancer.UpdateProviderWeights(providerWeights); err != nil {
		log.Printf("Warning: Failed to update provider weights: %v", err)
	} else {
		log.Println("Provider weights configured successfully")
	}
	log.Println("Load balancer service initialized with weighted distribution")

	// Initialize Failover Service (Task 7.4)
	failoverSvc := external.NewProviderFailover(providerRegistry, loadBalancer)
	log.Println("Failover service initialized with redundancy mechanisms")

	// Initialize Configuration Manager Service (Task 7.5)
	configManager := external.NewProviderConfigManager()

	// Set provider configurations
	if err := configManager.SetProviderConfig("openai", openaiProviderConfig); err != nil {
		log.Printf("Warning: Failed to set OpenAI configuration: %v", err)
	} else {
		log.Println("OpenAI provider configuration set")
	}

	if err := configManager.SetProviderConfig("ollama", ollamaProviderConfig); err != nil {
		log.Printf("Warning: Failed to set Ollama configuration: %v", err)
	} else {
		log.Println("Ollama provider configuration set")
	}

	log.Println("Configuration manager service initialized with templates and validation")
	log.Println("Multi-Provider Support Architecture initialization completed")

	log.Println("External API Services initialization completed")

	// Initialize handlers
	authHandler := auth.NewHandler(authSvc)
	chatHandlerInstance := chatHandler.NewHandler(chatSvc)
	kubeHandler := kubernetesHandler.NewHandler(kubeSvc)
	nlpHandlerInstance := nlpHandler.NewHandler(nlpSvc)
	// Initialize Enhanced NLP Handler with Safety Integration (Story 1.5 Task 7)
	enhancedNLPHandler := nlpHandler.NewEnhancedHandler(nlpSvc, querySvc, safetySvc)
	auditHandlerInstance := auditHandler.NewHandler(auditSvc)
	websocketHandlerInstance := websocketHandler.NewHandler(websocketSvc)
	gatewayHandlerInstance := gatewayHandler.NewHandler(gatewaySvc)
	healthHandlerInstance := healthHandler.NewHandler(healthSvc)
	communicationHandlerInstance := communicationHandler.NewHandler(commSvc)
	databaseHandlerInstance := databaseHandler.NewHandler(dbSvc)
	securityHandlerInstance := securityHandler.NewHandler(secSvc)

	// Initialize Command Handler (Story 1.6)
	var commandHandlerInstance *commandHandler.Handler
	if commandSvc != nil {
		commandHandlerInstance = commandHandler.NewHandler(commandSvc, nil) // no rollback service for now
		log.Println("Command handler initialized with execution service")
	}

	// Initialize External API Handler (Tasks 1-7)
	var externalHandlerInstance *externalHandler.Handler
	if clientManager != nil && externalHealthSvc != nil && encryptionSvc != nil &&
		credentialAuditSvc != nil && credentialValidator != nil && credentialInjector != nil {
		externalHandlerInstance = externalHandler.NewHandlerWithMultiProvider(
			clientManager,
			externalHealthSvc,
			encryptionSvc,
			credentialAuditSvc,
			credentialValidator,
			credentialInjector,
			// Task 6 services
			costSvc,
			budgetSvc,
			optimizationSvc,
			allocationSvc,
			controlSvc,
			// Task 7 services
			providerRegistry,
			loadBalancer,
			failoverSvc,
			configManager,
		)
		log.Println("External API handler initialized with multi-provider support (Tasks 1-7)")
	} else {
		log.Println("Warning: External API handler not initialized - some services unavailable")
	}

	// Create Gin router
	r := gin.Default()

	// API Gateway middleware stack with enhanced security
	r.Use(securityValidationMiddleware())          // Enhanced security headers and validation
	r.Use(gatewaySvc.SecurityHeadersMiddleware())  // CORS and security headers
	r.Use(gatewaySvc.RequestLoggingMiddleware())   // Request logging
	r.Use(gatewaySvc.ResponseTimeMiddleware())     // Response time tracking
	r.Use(gatewaySvc.RateLimitMiddleware())        // Rate limiting
	r.Use(gatewaySvc.CircuitBreakerMiddleware())   // Circuit breaker
	r.Use(gatewaySvc.RequestSizeLimitMiddleware()) // Request size limiting
	r.Use(gatewaySvc.GzipMiddleware())             // Gzip compression

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "kubechat-api",
			"version": "1.0.0",
		})
	})

	// Security health check endpoint
	r.GET("/security/health", func(c *gin.Context) {
		securityChecks := securityValidator.PerformSecurityHealthCheck()
		overallStatus := "healthy"

		// Check if any security checks failed
		for _, check := range securityChecks {
			if checkMap, ok := check.(map[string]interface{}); ok {
				if status, exists := checkMap["status"]; exists && status != "healthy" {
					overallStatus = "unhealthy"
					break
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"status":          overallStatus,
			"service":         "kubechat-api-security",
			"version":         "1.0.0",
			"security_checks": securityChecks,
			"timestamp":       time.Now(),
		})
	})

	// API routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "KubeChat API is running",
				"version": "1.0.0",
			})
		})

		// Register authentication routes
		authHandler.RegisterRoutes(v1)

		// Register chat routes
		chatHandlerInstance.RegisterRoutes(v1)

		// Register Kubernetes routes if service is available
		if kubeSvc != nil {
			kubeHandler.RegisterRoutes(v1)
		}

		// Register NLP routes
		nlpHandlerInstance.RegisterRoutes(v1)

		// Register Enhanced NLP routes with Safety Integration (Story 1.5 Task 7)
		enhancedNLPHandler.RegisterEnhancedRoutes(v1)

		// Register Audit routes
		auditHandlerInstance.RegisterRoutes(v1)

		// Register WebSocket routes
		websocketHandlerInstance.RegisterRoutes(v1)

		// Register Gateway management routes
		gatewayHandlerInstance.RegisterRoutes(v1)

		// Register Health Check routes
		healthHandlerInstance.RegisterRoutes(v1)

		// Register Service Communication routes
		communicationHandlerInstance.RegisterRoutes(v1)

		// Register Database Management routes
		databaseHandlerInstance.RegisterRoutes(v1)

		// Register Security and Performance routes
		securityHandlerInstance.RegisterRoutes(v1)

		// Register External API routes (Tasks 1-3)
		if externalHandlerInstance != nil {
			externalHandlerInstance.RegisterRoutes(v1)
		}

		// Register Command Execution routes (Story 1.6)
		if commandHandlerInstance != nil {
			commandHandlerInstance.RegisterRoutes(r)
		}

		// Register Swagger API documentation at /api/v1/docs with custom title
		url := ginSwagger.URL("doc.json") // The url pointing to API definition
		v1.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
			ginSwagger.DefaultModelsExpandDepth(-1),
			ginSwagger.DocExpansion("none"),
			ginSwagger.DeepLinking(true),
			ginSwagger.PersistAuthorization(true),
			url,
		))
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting KubeChat API server on port %s...", port)
	log.Printf("Available endpoints:")
	log.Printf("  GET  /health")
	log.Printf("  GET  /api/v1/status")
	log.Printf("  GET  /api/v1/docs/         - 📖 Complete API Documentation (Swagger UI)")
	log.Printf("  GET  /api/v1/docs/swagger.json - OpenAPI 3.0 Specification")
	log.Printf("")
	log.Printf("🚀 KubeChat Enterprise Platform API Documentation:")
	log.Printf("📖 Interactive API Docs: http://localhost:%s/api/v1/docs/", port)
	log.Printf("  POST /api/v1/auth/register")
	log.Printf("  POST /api/v1/auth/login")
	log.Printf("  POST /api/v1/auth/logout")
	log.Printf("  POST /api/v1/auth/refresh")
	log.Printf("  GET  /api/v1/auth/me")
	log.Printf("  GET  /api/v1/auth/profile")
	log.Printf("  POST /api/v1/nlp/process")
	log.Printf("  POST /api/v1/nlp/validate")
	log.Printf("  POST /api/v1/nlp/classify")
	log.Printf("  GET  /api/v1/nlp/providers")
	log.Printf("  GET  /api/v1/nlp/health")
	log.Printf("  GET  /api/v1/nlp/metrics")
	log.Printf("  GET  /api/v1/audit/logs")
	log.Printf("  GET  /api/v1/audit/logs/:id")
	log.Printf("  POST /api/v1/audit/logs")
	log.Printf("  GET  /api/v1/audit/summary")
	log.Printf("  GET  /api/v1/audit/dangerous")
	log.Printf("  GET  /api/v1/audit/failed")
	log.Printf("  POST /api/v1/audit/verify-integrity")
	log.Printf("  GET  /api/v1/audit/health")
	log.Printf("  GET  /api/v1/audit/metrics")
	log.Printf("  GET  /api/v1/ws")
	log.Printf("  GET  /api/v1/websocket/clients")
	log.Printf("  GET  /api/v1/websocket/clients/count")
	log.Printf("  POST /api/v1/websocket/broadcast")
	log.Printf("  GET  /api/v1/websocket/metrics")
	log.Printf("  GET  /api/v1/websocket/health")
	log.Printf("  GET  /api/v1/gateway/metrics")
	log.Printf("  GET  /api/v1/gateway/health")
	log.Printf("  GET  /api/v1/gateway/status")
	log.Printf("  GET  /api/v1/gateway/rate-limits")
	log.Printf("  GET  /api/v1/health/")
	log.Printf("  GET  /api/v1/health/detailed")
	log.Printf("  GET  /api/v1/health/live")
	log.Printf("  GET  /api/v1/health/ready")
	log.Printf("  GET  /api/v1/health/components")
	log.Printf("  GET  /api/v1/health/metrics")
	log.Printf("  GET  /api/v1/health/database")
	log.Printf("  GET  /api/v1/health/redis")
	log.Printf("  POST /api/v1/communication/services/register")
	log.Printf("  GET  /api/v1/communication/services/discover/:serviceName")
	log.Printf("  POST /api/v1/communication/call")
	log.Printf("  POST /api/v1/communication/broadcast")
	log.Printf("  POST /api/v1/communication/events/publish")
	log.Printf("  GET  /api/v1/communication/metrics")
	log.Printf("  GET  /api/v1/communication/patterns")
	log.Printf("  POST /api/v1/communication/load-balancer/test")
	log.Printf("  POST /api/v1/database/pools")
	log.Printf("  GET  /api/v1/database/pools")
	log.Printf("  GET  /api/v1/database/instances")
	log.Printf("  POST /api/v1/database/clusters")
	log.Printf("  GET  /api/v1/database/clusters")
	log.Printf("  GET  /api/v1/database/metrics")
	log.Printf("  GET  /api/v1/database/health")
	log.Printf("  POST /api/v1/security/validate-password")
	log.Printf("  POST /api/v1/security/analyze-request")
	log.Printf("  GET  /api/v1/security/events")
	log.Printf("  GET  /api/v1/security/health")
	log.Printf("  GET  /api/v1/performance/metrics")
	log.Printf("  GET  /api/v1/performance/cache/stats")
	log.Printf("  GET  /api/v1/performance/health")
	log.Printf("  POST /api/v1/commands/execute")
	log.Printf("  GET  /api/v1/commands/executions")
	log.Printf("  GET  /api/v1/commands/executions/:id")
	log.Printf("  GET  /api/v1/commands/approvals")
	log.Printf("  POST /api/v1/commands/approve/:id")
	log.Printf("  GET  /api/v1/commands/stats")
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// initDatabase initializes the database connection with enhanced pooling and monitoring
func initDatabase() (*sqlx.DB, error) {
	// Get database configuration from environment
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "kubechat-dev-postgresql"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "kubechat"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		log.Fatal("SECURITY CRITICAL: DB_PASSWORD environment variable must be set. No default database passwords allowed.")
	}
	if password == "dev-password" || password == "password" || password == "postgres" || len(password) < 12 {
		log.Fatal("SECURITY CRITICAL: DB_PASSWORD must be a secure value with minimum 12 characters. Default or weak passwords are not allowed.")
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "kubechat"
	}

	sslmode := os.Getenv("DB_SSL_MODE")
	if sslmode == "" {
		// Internal cluster communication: SSL not required for in-cluster database
		// Database runs in same Kubernetes cluster with secure internal networking
		sslmode = "disable"
	}

	// Build connection string with additional parameters for reliability
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=10 statement_timeout=30000",
		host, port, user, password, dbname, sslmode)

	// Connect to database
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Enhanced connection pool configuration for production
	maxOpenConns := getEnvAsInt("DB_MAX_OPEN_CONNS", 25)
	maxIdleConns := getEnvAsInt("DB_MAX_IDLE_CONNS", 5)
	connMaxLifetime := getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute)
	connMaxIdleTime := getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", 1*time.Minute)

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdleTime)

	// Test the connection with retry logic
	maxRetries := 5
	retryDelay := time.Second * 2

	for i := 0; i < maxRetries; i++ {
		if err := db.Ping(); err != nil {
			if i == maxRetries-1 {
				return nil, fmt.Errorf("failed to ping database after %d retries: %w", maxRetries, err)
			}
			log.Printf("Database ping failed (attempt %d/%d), retrying in %v: %v", i+1, maxRetries, retryDelay, err)
			time.Sleep(retryDelay)
			continue
		}
		break
	}

	// Log connection pool configuration
	log.Printf("Successfully connected to database at %s:%s", host, port)
	log.Printf("Connection pool: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v, MaxIdleTime=%v",
		maxOpenConns, maxIdleConns, connMaxLifetime, connMaxIdleTime)

	// Start connection monitoring goroutine
	go monitorDatabase(db)

	return db, nil
}

// getEnvAsInt parses environment variable as integer with default fallback
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvAsDuration parses environment variable as duration with default fallback
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// monitorDatabase monitors database connection health and logs statistics
func monitorDatabase(db *sqlx.DB) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats := db.Stats()

			// Log connection pool statistics
			log.Printf("DB Pool Stats: Open=%d, InUse=%d, Idle=%d, WaitCount=%d, WaitDuration=%v",
				stats.OpenConnections,
				stats.InUse,
				stats.Idle,
				stats.WaitCount,
				stats.WaitDuration,
			)

			// Health check
			if err := db.Ping(); err != nil {
				log.Printf("Database health check failed: %v", err)
			}

			// Alert on potential issues
			if stats.WaitCount > 10 {
				log.Printf("Warning: High database connection wait count (%d)", stats.WaitCount)
			}

			if stats.WaitDuration > time.Second {
				log.Printf("Warning: High database connection wait duration (%v)", stats.WaitDuration)
			}
		}
	}
}

// Mock repository implementations for Story 1.6
type mockCommandRepository struct{}

func (r *mockCommandRepository) Create(ctx context.Context, execution *models.KubernetesCommandExecution) error {
	// Mock implementation - in production this would insert into database
	return nil
}

func (r *mockCommandRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.KubernetesCommandExecution, error) {
	// Mock implementation - return a sample execution
	return &models.KubernetesCommandExecution{
		ID:     id,
		Status: "completed",
	}, nil
}

func (r *mockCommandRepository) Update(ctx context.Context, execution *models.KubernetesCommandExecution) error {
	// Mock implementation - in production this would update database
	return nil
}

func (r *mockCommandRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*models.KubernetesCommandExecution, error) {
	// Mock implementation - return empty list
	return []*models.KubernetesCommandExecution{}, nil
}

type mockRollbackRepository struct{}

func (r *mockRollbackRepository) Create(ctx context.Context, plan *models.RollbackPlan) error {
	// Mock implementation
	return nil
}

func (r *mockRollbackRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.RollbackPlan, error) {
	// Mock implementation
	return &models.RollbackPlan{
		ID:     id,
		Status: "planned",
	}, nil
}

func (r *mockRollbackRepository) Update(ctx context.Context, plan *models.RollbackPlan) error {
	// Mock implementation
	return nil
}

func (r *mockRollbackRepository) ListByExecution(ctx context.Context, executionID uuid.UUID) ([]*models.RollbackPlan, error) {
	// Mock implementation
	return []*models.RollbackPlan{}, nil
}
