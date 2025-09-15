package user

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// Service defines the interface for user operations
type Service interface {
	CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) (*models.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, offset, limit int) ([]*models.User, error)
	AuthenticateUser(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error)
	CreateUserSession(ctx context.Context, userID uuid.UUID, ipAddress, userAgent *string) (*models.UserSession, error)
	ValidateSession(ctx context.Context, sessionToken string) (*models.User, error)
	LogoutUser(ctx context.Context, sessionToken string) error
}

// Repository defines the interface for user data access
type Repository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*models.User, error)
	CreateSession(ctx context.Context, session *models.UserSession) error
	GetSessionByToken(ctx context.Context, token string) (*models.UserSession, error)
	DeleteSession(ctx context.Context, token string) error
}

// service implements the Service interface
type service struct {
	repo Repository
}

// New creates a new user service instance
func New(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// CreateUser creates a new user in the system
func (s *service) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.User, error) {
	// Validate role
	if req.Role != "" && !models.ValidateRole(req.Role) {
		return nil, fmt.Errorf("invalid role: %s", req.Role)
	}

	// Set default role if not specified
	role := req.Role
	if role == "" {
		role = models.RoleUser
	}

	// Check if username already exists
	existingUser, err := s.repo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("username already exists: %s", req.Username)
	}

	// Check if email already exists
	existingUser, err = s.repo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("email already exists: %s", req.Email)
	}

	// Create new user
	user := &models.User{
		ID:        uuid.New(),
		Username:  req.Username,
		Email:     req.Email,
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Hash password
	if err := user.HashPassword(req.Password); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Save to repository
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by their ID
func (s *service) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return user, nil
}

// GetUserByUsername retrieves a user by their username
func (s *service) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return user, nil
}

// GetUserByEmail retrieves a user by their email
func (s *service) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

// UpdateUser updates an existing user
func (s *service) UpdateUser(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) (*models.User, error) {
	// Get existing user
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Update fields if provided
	if req.Username != "" {
		// Check if new username already exists (and it's not the current user)
		existingUser, err := s.repo.GetByUsername(ctx, req.Username)
		if err == nil && existingUser != nil && existingUser.ID != id {
			return nil, fmt.Errorf("username already exists: %s", req.Username)
		}
		user.Username = req.Username
	}

	if req.Email != "" {
		// Check if new email already exists (and it's not the current user)
		existingUser, err := s.repo.GetByEmail(ctx, req.Email)
		if err == nil && existingUser != nil && existingUser.ID != id {
			return nil, fmt.Errorf("email already exists: %s", req.Email)
		}
		user.Email = req.Email
	}

	user.UpdatedAt = time.Now()

	// Save changes
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUser deletes a user from the system
func (s *service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// Check if user exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Delete user
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListUsers retrieves a paginated list of users
func (s *service) ListUsers(ctx context.Context, offset, limit int) ([]*models.User, error) {
	// Set sensible defaults
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	users, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

// AuthenticateUser authenticates a user and creates a session
func (s *service) AuthenticateUser(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
	// Get user by username
	user, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: invalid username or password")
	}

	// Check password
	if err := user.CheckPassword(req.Password); err != nil {
		return nil, fmt.Errorf("authentication failed: invalid username or password")
	}

	// Create new session
	session, err := models.NewUserSession(user.ID, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Save session
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return &models.LoginResponse{
		User:         *user,
		SessionToken: session.SessionToken,
		ExpiresAt:    session.ExpiresAt,
	}, nil
}

// CreateUserSession creates a new session for a user
func (s *service) CreateUserSession(ctx context.Context, userID uuid.UUID, ipAddress, userAgent *string) (*models.UserSession, error) {
	// Verify user exists
	_, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Create session
	session, err := models.NewUserSession(userID, ipAddress, userAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Save session
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return session, nil
}

// ValidateSession validates a session token and returns the associated user
func (s *service) ValidateSession(ctx context.Context, sessionToken string) (*models.User, error) {
	// Get session by token
	session, err := s.repo.GetSessionByToken(ctx, sessionToken)
	if err != nil {
		return nil, fmt.Errorf("invalid session token")
	}

	// Check if session is expired
	if session.IsExpired() {
		// Clean up expired session
		_ = s.repo.DeleteSession(ctx, sessionToken)
		return nil, fmt.Errorf("session expired")
	}

	// Get user
	user, err := s.repo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found for session")
	}

	return user, nil
}

// LogoutUser invalidates a user session
func (s *service) LogoutUser(ctx context.Context, sessionToken string) error {
	// Delete session
	if err := s.repo.DeleteSession(ctx, sessionToken); err != nil {
		return fmt.Errorf("failed to logout user: %w", err)
	}

	return nil
}
