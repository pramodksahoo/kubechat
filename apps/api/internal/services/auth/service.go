package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/repositories"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrSessionExpired     = errors.New("session expired")
	ErrInvalidToken       = errors.New("invalid token")
	ErrUserAlreadyExists  = errors.New("user already exists")
)

// Claims represents JWT claims for authentication
type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	SessionID uuid.UUID `json:"session_id"`
	jwt.RegisteredClaims
}

// Service defines the authentication service interface
type Service interface {
	// User management
	RegisterUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error)
	LoginUser(ctx context.Context, req *models.LoginRequest, ipAddress, userAgent *string) (*models.LoginResponse, error)
	LogoutUser(ctx context.Context, sessionToken string) error

	// Token and session management
	GenerateJWT(user *models.User, sessionID uuid.UUID) (string, error)
	ValidateJWT(tokenString string) (*Claims, error)
	RefreshSession(ctx context.Context, sessionToken string) (*models.LoginResponse, error)

	// User operations
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetCurrentUser(ctx context.Context, sessionToken string) (*models.User, error)

	// Session management
	ValidateSession(ctx context.Context, sessionToken string) (*models.UserSession, error)
	CleanupExpiredSessions(ctx context.Context) error
}

// service implements the authentication service
type service struct {
	userRepo  repositories.UserRepository
	jwtSecret string
	jwtExpiry time.Duration
}

// NewService creates a new authentication service
func NewService(userRepo repositories.UserRepository, jwtSecret string) Service {
	return &service{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtExpiry: 24 * time.Hour, // JWT expires in 24 hours
	}
}

// RegisterUser creates a new user account
func (s *service) RegisterUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	// Check if user already exists by username
	existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Check if user already exists by email
	existingUser, err = s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Create new user
	user := &models.User{
		ID:        uuid.New(),
		Username:  req.Username,
		Email:     req.Email,
		Role:      req.Role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set default role if not provided
	if user.Role == "" {
		user.Role = models.RoleUser
	}

	// Validate role
	if !models.ValidateRole(user.Role) {
		user.Role = models.RoleUser
	}

	// Hash password
	if err := user.HashPassword(req.Password); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Save user to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// LoginUser authenticates a user and creates a session
func (s *service) LoginUser(ctx context.Context, req *models.LoginRequest, ipAddress, userAgent *string) (*models.LoginResponse, error) {
	// Get user by username
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := user.CheckPassword(req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Create new session
	session, err := models.NewUserSession(user.ID, ipAddress, userAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Save session to database
	if err := s.userRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	// Generate JWT token
	jwtToken, err := s.GenerateJWT(user, session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	return &models.LoginResponse{
		User:         *user,
		SessionToken: jwtToken,
		ExpiresAt:    session.ExpiresAt,
	}, nil
}

// LogoutUser invalidates a user session
func (s *service) LogoutUser(ctx context.Context, sessionToken string) error {
	// Validate and get session
	session, err := s.ValidateSession(ctx, sessionToken)
	if err != nil {
		return err
	}

	// Delete session from database
	return s.userRepo.DeleteSession(ctx, session.ID)
}

// GenerateJWT creates a JWT token for authenticated user
func (s *service) GenerateJWT(user *models.User, sessionID uuid.UUID) (string, error) {
	claims := Claims{
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "kubechat-auth",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// ValidateJWT validates and parses a JWT token
func (s *service) ValidateJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshSession extends an existing session
func (s *service) RefreshSession(ctx context.Context, sessionToken string) (*models.LoginResponse, error) {
	// Validate current session
	session, err := s.ValidateSession(ctx, sessionToken)
	if err != nil {
		return nil, err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Extend session
	session.Extend(models.DefaultSessionExpiry)

	// Generate new JWT token
	jwtToken, err := s.GenerateJWT(user, session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	return &models.LoginResponse{
		User:         *user,
		SessionToken: jwtToken,
		ExpiresAt:    session.ExpiresAt,
	}, nil
}

// GetUserByID retrieves a user by ID
func (s *service) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// GetCurrentUser retrieves the current authenticated user
func (s *service) GetCurrentUser(ctx context.Context, sessionToken string) (*models.User, error) {
	session, err := s.ValidateSession(ctx, sessionToken)
	if err != nil {
		return nil, err
	}

	return s.userRepo.GetByID(ctx, session.UserID)
}

// ValidateSession validates a session token and returns the session
func (s *service) ValidateSession(ctx context.Context, sessionToken string) (*models.UserSession, error) {
	// Parse JWT token to get session ID
	claims, err := s.ValidateJWT(sessionToken)
	if err != nil {
		return nil, err
	}

	// Get session from database
	session, err := s.userRepo.GetSessionByToken(ctx, sessionToken)
	if err != nil || session == nil {
		return nil, ErrInvalidToken
	}

	// Check if session is expired
	if session.IsExpired() {
		// Clean up expired session
		s.userRepo.DeleteSession(ctx, session.ID)
		return nil, ErrSessionExpired
	}

	// Verify session ID matches JWT claims
	if session.ID != claims.SessionID {
		return nil, ErrInvalidToken
	}

	return session, nil
}

// CleanupExpiredSessions removes expired sessions from database
func (s *service) CleanupExpiredSessions(ctx context.Context) error {
	return s.userRepo.DeleteExpiredSessions(ctx)
}
