package websocket

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/audit"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/auth"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/kubernetes"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/nlp"
)

// Service defines the interface for WebSocket service
type Service interface {
	// HandleWebSocket handles WebSocket connection upgrade and message processing
	HandleWebSocket(c *gin.Context)

	// BroadcastToAll broadcasts a message to all connected clients
	BroadcastToAll(message *models.WebSocketMessage) error

	// BroadcastToUser broadcasts a message to specific user
	BroadcastToUser(userID uuid.UUID, message *models.WebSocketMessage) error

	// BroadcastToSubscribers broadcasts to clients subscribed to specific topics
	BroadcastToSubscribers(topics []string, message *models.WebSocketMessage) error

	// GetConnectedClients returns list of connected clients
	GetConnectedClients() []*models.WebSocketClient

	// GetClientCount returns number of connected clients
	GetClientCount() int

	// GetMetrics returns WebSocket service metrics
	GetMetrics() *models.WebSocketMetrics

	// DisconnectClient forcefully disconnects a client
	DisconnectClient(clientID string) error

	// HealthCheck performs health check
	HealthCheck(ctx context.Context) error

	// Shutdown gracefully shuts down the service
	Shutdown(ctx context.Context) error
}

// Config represents WebSocket service configuration
type Config struct {
	// Connection settings
	ReadTimeout    time.Duration `json:"read_timeout"`
	WriteTimeout   time.Duration `json:"write_timeout"`
	PingPeriod     time.Duration `json:"ping_period"`
	PongWait       time.Duration `json:"pong_wait"`
	MaxMessageSize int64         `json:"max_message_size"`

	// Buffer settings
	ReadBufferSize  int `json:"read_buffer_size"`
	WriteBufferSize int `json:"write_buffer_size"`

	// Client management
	MaxClients        int           `json:"max_clients"`
	ClientTimeout     time.Duration `json:"client_timeout"`
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`

	// Command execution
	MaxConcurrentCommands int           `json:"max_concurrent_commands"`
	CommandTimeout        time.Duration `json:"command_timeout"`

	// Security
	AllowedOrigins []string `json:"allowed_origins"`
	RequireAuth    bool     `json:"require_auth"`

	// Logging
	EnableDebugLogging bool   `json:"enable_debug_logging"`
	LogLevel           string `json:"log_level"`
}

// client represents a WebSocket client connection
type client struct {
	// Connection details
	conn   *websocket.Conn
	send   chan *models.WebSocketMessage
	hub    *hub
	logger *slog.Logger

	// Client information
	client *models.WebSocketClient

	// Context and cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Execution tracking
	activeCommands map[string]*models.CommandExecution
	commandsMutex  sync.RWMutex
}

// hub manages WebSocket connections and message routing
type hub struct {
	// Client management
	clients    map[string]*client
	register   chan *client
	unregister chan *client
	broadcast  chan *models.WebSocketMessage
	mutex      sync.RWMutex

	// Message routing
	userClients      map[uuid.UUID][]*client
	topicSubscribers map[string][]*client

	// Services
	authService       auth.Service
	kubernetesService kubernetes.Service
	nlpService        nlp.Service
	auditService      audit.Service

	// Configuration and state
	config    *Config
	logger    *slog.Logger
	metrics   *models.WebSocketMetrics
	startTime time.Time

	// Graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// service implements the Service interface
type service struct {
	hub      *hub
	upgrader websocket.Upgrader
	config   *Config
	logger   *slog.Logger
}

// NewService creates a new WebSocket service
func NewService(
	authService auth.Service,
	kubernetesService kubernetes.Service,
	nlpService nlp.Service,
	auditService audit.Service,
	config *Config,
) Service {
	if config == nil {
		config = &Config{
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
			AllowedOrigins:        []string{"*"},
			RequireAuth:           true,
			EnableDebugLogging:    false,
			LogLevel:              "info",
		}
	}

	// Initialize logger
	logLevel := slog.LevelInfo
	switch config.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	// Create hub
	ctx, cancel := context.WithCancel(context.Background())
	h := &hub{
		clients:           make(map[string]*client),
		register:          make(chan *client, 256),
		unregister:        make(chan *client, 256),
		broadcast:         make(chan *models.WebSocketMessage, 1024),
		userClients:       make(map[uuid.UUID][]*client),
		topicSubscribers:  make(map[string][]*client),
		authService:       authService,
		kubernetesService: kubernetesService,
		nlpService:        nlpService,
		auditService:      auditService,
		config:            config,
		logger:            logger,
		metrics: &models.WebSocketMetrics{
			ConnectedClients:  0,
			TotalConnections:  0,
			ActiveCommands:    0,
			CompletedCommands: 0,
			FailedCommands:    0,
			MessagesSent:      0,
			MessagesReceived:  0,
			ErrorCount:        0,
		},
		startTime: time.Now(),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Configure WebSocket upgrader
	upgrader := websocket.Upgrader{
		ReadBufferSize:  config.ReadBufferSize,
		WriteBufferSize: config.WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if len(config.AllowedOrigins) == 0 {
				return true
			}
			for _, allowedOrigin := range config.AllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					return true
				}
			}
			return false
		},
		Subprotocols: []string{"kubechat-ws"},
	}

	svc := &service{
		hub:      h,
		upgrader: upgrader,
		config:   config,
		logger:   logger,
	}

	// Start hub
	h.wg.Add(1)
	go h.run()

	logger.Info("WebSocket service initialized",
		"max_clients", config.MaxClients,
		"require_auth", config.RequireAuth,
		"ping_period", config.PingPeriod,
		"max_message_size", config.MaxMessageSize)

	return svc
}

// HandleWebSocket handles WebSocket connection upgrade and client management
func (s *service) HandleWebSocket(c *gin.Context) {
	// Upgrade HTTP connection to WebSocket
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.logger.Error("Failed to upgrade WebSocket connection", "error", err)
		return
	}

	// Check client limit
	if s.hub.getClientCount() >= s.config.MaxClients {
		s.logger.Warn("Max clients reached, rejecting connection")
		conn.Close()
		return
	}

	// Extract client info from request
	clientID := uuid.New().String()
	ipAddress := extractIPAddress(c.Request)
	userAgent := c.Request.UserAgent()

	s.logger.Info("New WebSocket connection",
		"client_id", clientID,
		"ip_address", ipAddress,
		"user_agent", userAgent)

	// Create client
	ctx, cancel := context.WithCancel(s.hub.ctx)
	client := &client{
		conn:   conn,
		send:   make(chan *models.WebSocketMessage, 256),
		hub:    s.hub,
		logger: s.logger,
		client: &models.WebSocketClient{
			ID:             clientID,
			ConnectedAt:    time.Now(),
			LastPing:       time.Now(),
			IPAddress:      ipAddress,
			UserAgent:      userAgent,
			Subscriptions:  []string{},
			ActiveCommands: []string{},
		},
		ctx:            ctx,
		cancel:         cancel,
		activeCommands: make(map[string]*models.CommandExecution),
	}

	// Configure connection
	conn.SetReadLimit(s.config.MaxMessageSize)
	conn.SetReadDeadline(time.Now().Add(s.config.PongWait))
	conn.SetPongHandler(func(string) error {
		client.client.LastPing = time.Now()
		conn.SetReadDeadline(time.Now().Add(s.config.PongWait))
		return nil
	})

	// Register client
	s.hub.register <- client

	// Start goroutines for this client
	go client.readMessages()
	go client.writeMessages()
}

// BroadcastToAll broadcasts a message to all connected clients
func (s *service) BroadcastToAll(message *models.WebSocketMessage) error {
	select {
	case s.hub.broadcast <- message:
		return nil
	default:
		return fmt.Errorf("broadcast channel full")
	}
}

// BroadcastToUser broadcasts a message to specific user
func (s *service) BroadcastToUser(userID uuid.UUID, message *models.WebSocketMessage) error {
	s.hub.mutex.RLock()
	clients := s.hub.userClients[userID]
	s.hub.mutex.RUnlock()

	if len(clients) == 0 {
		return fmt.Errorf("no clients found for user %s", userID)
	}

	for _, client := range clients {
		select {
		case client.send <- message:
		default:
			s.logger.Warn("Client send channel full", "client_id", client.client.ID)
		}
	}

	return nil
}

// BroadcastToSubscribers broadcasts to clients subscribed to specific topics
func (s *service) BroadcastToSubscribers(topics []string, message *models.WebSocketMessage) error {
	s.hub.mutex.RLock()
	clientSet := make(map[string]*client)

	for _, topic := range topics {
		if subscribers, exists := s.hub.topicSubscribers[topic]; exists {
			for _, client := range subscribers {
				clientSet[client.client.ID] = client
			}
		}
	}
	s.hub.mutex.RUnlock()

	if len(clientSet) == 0 {
		return fmt.Errorf("no subscribers found for topics: %v", topics)
	}

	for _, client := range clientSet {
		select {
		case client.send <- message:
		default:
			s.logger.Warn("Client send channel full", "client_id", client.client.ID)
		}
	}

	return nil
}

// GetConnectedClients returns list of connected clients
func (s *service) GetConnectedClients() []*models.WebSocketClient {
	s.hub.mutex.RLock()
	defer s.hub.mutex.RUnlock()

	clients := make([]*models.WebSocketClient, 0, len(s.hub.clients))
	for _, client := range s.hub.clients {
		clients = append(clients, client.client)
	}

	return clients
}

// GetClientCount returns number of connected clients
func (s *service) GetClientCount() int {
	return s.hub.getClientCount()
}

// GetMetrics returns WebSocket service metrics
func (s *service) GetMetrics() *models.WebSocketMetrics {
	s.hub.mutex.RLock()
	defer s.hub.mutex.RUnlock()

	// Update current metrics
	metrics := *s.hub.metrics
	metrics.ConnectedClients = len(s.hub.clients)
	metrics.Uptime = time.Since(s.hub.startTime).String()

	return &metrics
}

// DisconnectClient forcefully disconnects a client
func (s *service) DisconnectClient(clientID string) error {
	s.hub.mutex.RLock()
	client, exists := s.hub.clients[clientID]
	s.hub.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("client %s not found", clientID)
	}

	client.cancel()
	return nil
}

// HealthCheck performs health check
func (s *service) HealthCheck(ctx context.Context) error {
	// Check if hub is running
	select {
	case <-s.hub.ctx.Done():
		return fmt.Errorf("websocket hub not running")
	default:
	}

	// Check service dependencies (simplified - auth service doesn't have HealthCheck method)
	// Just verify services are not nil
	if s.hub.authService == nil {
		return fmt.Errorf("auth service not available")
	}

	if s.hub.nlpService == nil {
		s.logger.Warn("NLP service not available")
	}

	if s.hub.auditService == nil {
		s.logger.Warn("Audit service not available")
	}

	return nil
}

// Shutdown gracefully shuts down the service
func (s *service) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down WebSocket service")

	s.hub.cancel()

	// Wait for hub to finish with timeout
	done := make(chan struct{})
	go func() {
		s.hub.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("WebSocket service shutdown completed")
		return nil
	case <-ctx.Done():
		s.logger.Warn("WebSocket service shutdown timed out")
		return ctx.Err()
	}
}

// Helper functions

// extractIPAddress extracts the real IP address from request
func extractIPAddress(req *http.Request) string {
	// Check for forwarded IP addresses
	if forwarded := req.Header.Get("X-Forwarded-For"); forwarded != "" {
		// Take the first IP if multiple are present
		if idx := strings.Index(forwarded, ","); idx != -1 {
			return strings.TrimSpace(forwarded[:idx])
		}
		return forwarded
	}

	if realIP := req.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	return req.RemoteAddr
}

// getClientCount returns the current client count
func (h *hub) getClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}
