package models_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

func TestUser_HashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "complex password",
			password: "MyVerySecureP@ssw0rd!",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false, // bcrypt can hash empty strings
		},
		{
			name:     "long password",
			password: "ThisIsAVeryLongPasswordThatExceeds72BytesWhichIsTheMaximumLengthForBcryptButItShouldStillWork",
			wantErr:  true, // bcrypt errors on passwords > 72 bytes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &models.User{}
			err := user.HashPassword(tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, user.PasswordHash)

			// Verify the hash is actually a bcrypt hash
			assert.True(t, len(user.PasswordHash) >= 60) // bcrypt hashes are at least 60 chars
			assert.True(t, user.PasswordHash[:4] == "$2a$" || user.PasswordHash[:4] == "$2b$")

			// Verify the cost is correct (14)
			cost, err := bcrypt.Cost([]byte(user.PasswordHash))
			require.NoError(t, err)
			assert.Equal(t, models.BcryptCost, cost)
		})
	}
}

func TestUser_CheckPassword(t *testing.T) {
	user := &models.User{}
	originalPassword := "testPassword123"

	// Hash the password first
	err := user.HashPassword(originalPassword)
	require.NoError(t, err)

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "correct password",
			password: originalPassword,
			wantErr:  false,
		},
		{
			name:     "incorrect password",
			password: "wrongPassword",
			wantErr:  true,
		},
		{
			name:     "empty password when hash exists",
			password: "",
			wantErr:  true,
		},
		{
			name:     "case sensitive",
			password: "TestPassword123", // Different case
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := user.CheckPassword(tt.password)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUser_RoleMethods(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		isAdmin  bool
		canWrite bool
		canRead  bool
	}{
		{
			name:     "admin role",
			role:     models.RoleAdmin,
			isAdmin:  true,
			canWrite: true,
			canRead:  true,
		},
		{
			name:     "user role",
			role:     models.RoleUser,
			isAdmin:  false,
			canWrite: true,
			canRead:  true,
		},
		{
			name:     "viewer role",
			role:     models.RoleViewer,
			isAdmin:  false,
			canWrite: false,
			canRead:  true,
		},
		{
			name:     "invalid role",
			role:     "invalid",
			isAdmin:  false,
			canWrite: false,
			canRead:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &models.User{Role: tt.role}

			assert.Equal(t, tt.isAdmin, user.IsAdmin())
			assert.Equal(t, tt.canWrite, user.CanWrite())
			assert.Equal(t, tt.canRead, user.CanRead())
		})
	}
}

func TestGenerateSecureToken(t *testing.T) {
	// Generate multiple tokens to ensure they're unique
	tokens := make(map[string]bool)

	for i := 0; i < 100; i++ {
		token, err := models.GenerateSecureToken()

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.True(t, len(token) > 0)

		// Ensure token is unique
		assert.False(t, tokens[token], "Token should be unique")
		tokens[token] = true

		// Token should be base64 URL encoded, so check basic format
		assert.NotContains(t, token, " ")
		assert.NotContains(t, token, "\n")
	}
}

func TestNewUserSession(t *testing.T) {
	userID := uuid.New()
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0"

	session, err := models.NewUserSession(userID, &ipAddress, &userAgent)

	require.NoError(t, err)
	require.NotNil(t, session)

	assert.NotEqual(t, uuid.Nil, session.ID)
	assert.Equal(t, userID, session.UserID)
	assert.NotEmpty(t, session.SessionToken)
	assert.Equal(t, &ipAddress, session.IPAddress)
	assert.Equal(t, &userAgent, session.UserAgent)

	// Check expiration is in the future
	assert.True(t, session.ExpiresAt.After(time.Now()))

	// Check expiration is within expected range (24 hours)
	expectedExpiry := time.Now().Add(models.DefaultSessionExpiry)
	assert.WithinDuration(t, expectedExpiry, session.ExpiresAt, time.Minute)

	// Test with nil values
	session2, err := models.NewUserSession(userID, nil, nil)
	require.NoError(t, err)
	assert.Nil(t, session2.IPAddress)
	assert.Nil(t, session2.UserAgent)
}

func TestUserSession_IsExpired(t *testing.T) {
	userID := uuid.New()

	// Create expired session
	expiredSession := &models.UserSession{
		ID:        uuid.New(),
		UserID:    userID,
		ExpiresAt: time.Now().Add(-time.Hour), // Expired 1 hour ago
	}

	// Create valid session
	validSession := &models.UserSession{
		ID:        uuid.New(),
		UserID:    userID,
		ExpiresAt: time.Now().Add(time.Hour), // Expires 1 hour from now
	}

	assert.True(t, expiredSession.IsExpired())
	assert.False(t, expiredSession.IsValid())

	assert.False(t, validSession.IsExpired())
	assert.True(t, validSession.IsValid())
}

func TestUserSession_Extend(t *testing.T) {
	userID := uuid.New()
	session := &models.UserSession{
		ID:        uuid.New(),
		UserID:    userID,
		ExpiresAt: time.Now().Add(time.Hour),
	}

	originalExpiry := session.ExpiresAt
	extension := 2 * time.Hour

	session.Extend(extension)

	// Should be extended from now, not from original expiry
	expectedExpiry := time.Now().Add(extension)
	assert.WithinDuration(t, expectedExpiry, session.ExpiresAt, time.Minute)
	assert.True(t, session.ExpiresAt.After(originalExpiry))
}

func TestValidateRole(t *testing.T) {
	validRoles := []string{
		models.RoleAdmin,
		models.RoleUser,
		models.RoleViewer,
	}

	invalidRoles := []string{
		"invalid",
		"administrator",
		"superuser",
		"",
		"ADMIN", // Case sensitive
	}

	for _, role := range validRoles {
		assert.True(t, models.ValidateRole(role), "Role %s should be valid", role)
	}

	for _, role := range invalidRoles {
		assert.False(t, models.ValidateRole(role), "Role %s should be invalid", role)
	}
}

func TestUser_SecurityCompliance(t *testing.T) {
	// Test that bcrypt cost meets security requirements
	assert.GreaterOrEqual(t, models.BcryptCost, 12, "Bcrypt cost should be at least 12 for security")

	// Test that session tokens are sufficiently long
	assert.GreaterOrEqual(t, models.SessionTokenLength, 32, "Session tokens should be at least 32 bytes")

	// Test that default session expiry is reasonable
	assert.GreaterOrEqual(t, models.DefaultSessionExpiry, time.Hour, "Session expiry should be at least 1 hour")
	assert.LessOrEqual(t, models.DefaultSessionExpiry, 7*24*time.Hour, "Session expiry should be at most 7 days")
}

func BenchmarkUser_HashPassword(b *testing.B) {
	user := &models.User{}
	password := "testPassword123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = user.HashPassword(password)
	}
}

func BenchmarkUser_CheckPassword(b *testing.B) {
	user := &models.User{}
	password := "testPassword123"

	// Pre-hash the password
	err := user.HashPassword(password)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = user.CheckPassword(password)
	}
}
