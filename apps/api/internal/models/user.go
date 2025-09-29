package models

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system with authentication capabilities
type User struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	Username            string     `json:"username" db:"username"`
	Email               string     `json:"email" db:"email"`
	PasswordHash        string     `json:"-" db:"password_hash"` // Never expose in JSON
	Role                string     `json:"role" db:"role"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
	LastLogin           *time.Time `json:"last_login,omitempty" db:"last_login"`
	IsActive            bool       `json:"is_active" db:"is_active"`
	FailedLoginAttempts int        `json:"failed_login_attempts" db:"failed_login_attempts"`
	AccountLockedUntil  *time.Time `json:"account_locked_until,omitempty" db:"account_locked_until"`
	PasswordChangedAt   time.Time  `json:"password_changed_at" db:"password_changed_at"`
	MustChangePassword  bool       `json:"must_change_password" db:"must_change_password"`
}

// UserSession represents a user session with security tracking
type UserSession struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	SessionToken string    `json:"session_token" db:"session_token"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	IPAddress    *string   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    *string   `json:"user_agent,omitempty" db:"user_agent"`
}

// CreateUserRequest represents the data needed to create a new user
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=255"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role" validate:"omitempty,oneof=admin user viewer"`
}

// LoginRequest represents user login credentials
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// UpdateUserRequest represents user update payload
type UpdateUserRequest struct {
	Username string `json:"username,omitempty" validate:"omitempty,min=3,max=255"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
}

// LoginResponse represents successful login response
type LoginResponse struct {
	User         User      `json:"user"`
	SessionToken string    `json:"session_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// UserRole constants for role-based access control
const (
	RoleAdmin             = "admin"
	RoleUser              = "user"
	RoleViewer            = "viewer"
	RoleAuditor           = "auditor"
	RoleComplianceOfficer = "compliance_officer"
)

// Security configuration constants
const (
	BcryptCost           = 14             // Cost factor for bcrypt (2^14 iterations)
	SessionTokenLength   = 32             // Length of session tokens in bytes
	DefaultSessionExpiry = 24 * time.Hour // Default session expiration
)

// HashPassword securely hashes a password using bcrypt with high cost factor
func (u *User) HashPassword(password string) error {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedBytes)
	return nil
}

// CheckPassword verifies if the provided password matches the stored hash
func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}

// IsAdmin returns true if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// CanWrite returns true if the user can perform write operations
func (u *User) CanWrite() bool {
	return u.Role == RoleAdmin || u.Role == RoleUser
}

// CanRead returns true if the user can perform read operations
func (u *User) CanRead() bool {
	return u.Role == RoleAdmin || u.Role == RoleUser || u.Role == RoleViewer
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken() (string, error) {
	bytes := make([]byte, SessionTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// NewUserSession creates a new user session with secure token
func NewUserSession(userID uuid.UUID, ipAddress, userAgent *string) (*UserSession, error) {
	token, err := GenerateSecureToken()
	if err != nil {
		return nil, err
	}

	return &UserSession{
		ID:           uuid.New(),
		UserID:       userID,
		SessionToken: token,
		ExpiresAt:    time.Now().Add(DefaultSessionExpiry),
		CreatedAt:    time.Now(),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}, nil
}

// IsExpired returns true if the session has expired
func (s *UserSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid checks if the session is valid (not expired)
func (s *UserSession) IsValid() bool {
	return !s.IsExpired()
}

// Extend extends the session expiration time
func (s *UserSession) Extend(duration time.Duration) {
	s.ExpiresAt = time.Now().Add(duration)
}

// ValidateRole validates if the provided role is valid
func ValidateRole(role string) bool {
	switch role {
	case RoleAdmin, RoleUser, RoleViewer, RoleAuditor, RoleComplianceOfficer:
		return true
	default:
		return false
	}
}

type UserListFilters struct {
	Role   string
	Status string
	Search string
	Limit  int
	Offset int
}

type AdminCreateUserRequest struct {
	Username              string `json:"username"`
	Email                 string `json:"email"`
	Password              string `json:"password"`
	Role                  string `json:"role"`
	RequirePasswordChange bool   `json:"require_password_change"`
}

type AdminUpdateUserRequest struct {
	Email         *string `json:"email"`
	Role          *string `json:"role"`
	IsActive      *bool   `json:"is_active"`
	AccountLocked *bool   `json:"account_locked"`
	ResetPassword bool    `json:"reset_password"`
	NewPassword   *string `json:"new_password"`
}
