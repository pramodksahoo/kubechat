package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_HashPassword(t *testing.T) {
	user := &User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}

	password := "securepassword123"
	err := user.HashPassword(password)
	require.NoError(t, err)

	// Should have set password hash
	assert.NotEmpty(t, user.PasswordHash)
	assert.NotEqual(t, password, user.PasswordHash)

	// Should be able to verify password
	err = user.CheckPassword(password)
	assert.NoError(t, err)

	// Should fail with wrong password
	err = user.CheckPassword("wrongpassword")
	assert.Error(t, err)
}

func TestUser_CheckPassword(t *testing.T) {
	user := &User{}
	password := "testpassword123"

	// Hash the password first
	err := user.HashPassword(password)
	require.NoError(t, err)

	// Check correct password
	err = user.CheckPassword(password)
	assert.NoError(t, err)

	// Check incorrect password
	err = user.CheckPassword("wrongpassword")
	assert.Error(t, err)

	// Check empty password
	err = user.CheckPassword("")
	assert.Error(t, err)
}

func TestUser_IsAdmin(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"admin role", RoleAdmin, true},
		{"user role", RoleUser, false},
		{"viewer role", RoleViewer, false},
		{"empty role", "", false},
		{"invalid role", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			result := user.IsAdmin()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUser_CanWrite(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"admin role", RoleAdmin, true},
		{"user role", RoleUser, true},
		{"viewer role", RoleViewer, false},
		{"empty role", "", false},
		{"invalid role", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			result := user.CanWrite()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUser_CanRead(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"admin role", RoleAdmin, true},
		{"user role", RoleUser, true},
		{"viewer role", RoleViewer, true},
		{"empty role", "", false},
		{"invalid role", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			result := user.CanRead()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateSecureToken(t *testing.T) {
	token1, err := GenerateSecureToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token1)

	token2, err := GenerateSecureToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token2)

	// Tokens should be different
	assert.NotEqual(t, token1, token2)
}

func TestNewUserSession(t *testing.T) {
	userID := uuid.New()
	ipAddress := "192.168.1.1"
	userAgent := "Test Agent"

	session, err := NewUserSession(userID, &ipAddress, &userAgent)
	require.NoError(t, err)

	assert.Equal(t, userID, session.UserID)
	assert.Equal(t, &ipAddress, session.IPAddress)
	assert.Equal(t, &userAgent, session.UserAgent)
	assert.NotEmpty(t, session.SessionToken)
	assert.True(t, session.ExpiresAt.After(time.Now()))
	assert.False(t, session.IsExpired())
	assert.True(t, session.IsValid())
}

func TestUserSession_IsExpired(t *testing.T) {
	session := &UserSession{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		SessionToken: "test-token",
		ExpiresAt:    time.Now().Add(-1 * time.Hour), // Expired
		CreatedAt:    time.Now().Add(-2 * time.Hour),
	}

	assert.True(t, session.IsExpired())
	assert.False(t, session.IsValid())
}

func TestUserSession_IsValid(t *testing.T) {
	session := &UserSession{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		SessionToken: "test-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour), // Valid
		CreatedAt:    time.Now(),
	}

	assert.False(t, session.IsExpired())
	assert.True(t, session.IsValid())
}

func TestUserSession_Extend(t *testing.T) {
	session := &UserSession{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		SessionToken: "test-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		CreatedAt:    time.Now(),
	}

	originalExpiry := session.ExpiresAt
	session.Extend(24 * time.Hour)

	assert.True(t, session.ExpiresAt.After(originalExpiry))
}

func TestValidateRole(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"admin role", RoleAdmin, true},
		{"user role", RoleUser, true},
		{"viewer role", RoleViewer, true},
		{"empty role", "", false},
		{"invalid role", "invalid", false},
		{"case sensitive", "Admin", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateRole(tt.role)
			assert.Equal(t, tt.expected, result)
		})
	}
}
