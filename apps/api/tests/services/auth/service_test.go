package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/auth"
)

// MockUserRepository implements repositories.UserRepository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) CreateSession(ctx context.Context, session *models.UserSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockUserRepository) GetSessionByToken(ctx context.Context, token string) (*models.UserSession, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserSession), args.Error(1)
}

func (m *MockUserRepository) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteExpiredSessions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*models.UserSession, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.UserSession), args.Error(1)
}

func TestAuthService_RegisterUser(t *testing.T) {
	tests := []struct {
		name          string
		request       *models.CreateUserRequest
		mockSetup     func(*MockUserRepository)
		expectedError error
		expectedUser  bool
	}{
		{
			name: "successful registration",
			request: &models.CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
				Role:     "user",
			},
			mockSetup: func(repo *MockUserRepository) {
				// No existing user found
				repo.On("GetByUsername", mock.Anything, "testuser").Return(nil, errors.New("not found"))
				repo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, errors.New("not found"))
				repo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectedError: nil,
			expectedUser:  true,
		},
		{
			name: "username already exists",
			request: &models.CreateUserRequest{
				Username: "existinguser",
				Email:    "test@example.com",
				Password: "password123",
				Role:     "user",
			},
			mockSetup: func(repo *MockUserRepository) {
				existingUser := &models.User{Username: "existinguser"}
				repo.On("GetByUsername", mock.Anything, "existinguser").Return(existingUser, nil)
			},
			expectedError: auth.ErrUserAlreadyExists,
			expectedUser:  false,
		},
		{
			name: "email already exists",
			request: &models.CreateUserRequest{
				Username: "testuser",
				Email:    "existing@example.com",
				Password: "password123",
				Role:     "user",
			},
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetByUsername", mock.Anything, "testuser").Return(nil, errors.New("not found"))
				existingUser := &models.User{Email: "existing@example.com"}
				repo.On("GetByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)
			},
			expectedError: auth.ErrUserAlreadyExists,
			expectedUser:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			service := auth.NewService(mockRepo, "test-secret")

			user, err := service.RegisterUser(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				if tt.expectedUser {
					assert.NotNil(t, user)
					assert.Equal(t, tt.request.Username, user.Username)
					assert.Equal(t, tt.request.Email, user.Email)
					assert.NotEmpty(t, user.PasswordHash)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_LoginUser(t *testing.T) {
	testUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}
	testUser.HashPassword("password123")

	tests := []struct {
		name          string
		request       *models.LoginRequest
		mockSetup     func(*MockUserRepository)
		expectedError error
	}{
		{
			name: "successful login",
			request: &models.LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetByUsername", mock.Anything, "testuser").Return(testUser, nil)
				repo.On("CreateSession", mock.Anything, mock.AnythingOfType("*models.UserSession")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "user not found",
			request: &models.LoginRequest{
				Username: "nonexistent",
				Password: "password123",
			},
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetByUsername", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: auth.ErrInvalidCredentials,
		},
		{
			name: "wrong password",
			request: &models.LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetByUsername", mock.Anything, "testuser").Return(testUser, nil)
			},
			expectedError: auth.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			service := auth.NewService(mockRepo, "test-secret")

			clientIP := "127.0.0.1"
			userAgent := "test-agent"
			response, err := service.LoginUser(context.Background(), tt.request, &clientIP, &userAgent)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, testUser.Username, response.User.Username)
				assert.NotEmpty(t, response.SessionToken)
				assert.True(t, response.ExpiresAt.After(time.Now()))
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_ValidateJWT(t *testing.T) {
	testUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Role:     models.RoleUser,
	}
	sessionID := uuid.New()

	mockRepo := new(MockUserRepository)
	service := auth.NewService(mockRepo, "test-secret")

	// Generate valid JWT
	validToken, err := service.GenerateJWT(testUser, sessionID)
	require.NoError(t, err)

	tests := []struct {
		name          string
		token         string
		expectedError error
	}{
		{
			name:          "valid token",
			token:         validToken,
			expectedError: nil,
		},
		{
			name:          "invalid token",
			token:         "invalid.token.here",
			expectedError: auth.ErrInvalidToken,
		},
		{
			name:          "empty token",
			token:         "",
			expectedError: auth.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.ValidateJWT(tt.token)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, testUser.ID, claims.UserID)
				assert.Equal(t, testUser.Username, claims.Username)
				assert.Equal(t, testUser.Role, claims.Role)
				assert.Equal(t, sessionID, claims.SessionID)
			}
		})
	}
}

func TestAuthService_ValidateSession(t *testing.T) {
	testUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Role:     models.RoleUser,
	}
	sessionID := uuid.New()

	mockRepo := new(MockUserRepository)
	service := auth.NewService(mockRepo, "test-secret")

	// Generate valid JWT
	validToken, err := service.GenerateJWT(testUser, sessionID)
	require.NoError(t, err)

	validSession := &models.UserSession{
		ID:           sessionID,
		UserID:       testUser.ID,
		SessionToken: validToken,
		ExpiresAt:    time.Now().Add(time.Hour),
		CreatedAt:    time.Now(),
	}

	expiredSession := &models.UserSession{
		ID:           sessionID,
		UserID:       testUser.ID,
		SessionToken: validToken,
		ExpiresAt:    time.Now().Add(-time.Hour), // Expired
		CreatedAt:    time.Now(),
	}

	tests := []struct {
		name          string
		token         string
		mockSetup     func(*MockUserRepository)
		expectedError error
	}{
		{
			name:  "valid session",
			token: validToken,
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetSessionByToken", mock.Anything, validToken).Return(validSession, nil)
			},
			expectedError: nil,
		},
		{
			name:  "session not found",
			token: validToken,
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetSessionByToken", mock.Anything, validToken).Return(nil, errors.New("not found"))
			},
			expectedError: auth.ErrInvalidToken,
		},
		{
			name:  "expired session",
			token: validToken,
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetSessionByToken", mock.Anything, validToken).Return(expiredSession, nil)
				repo.On("DeleteSession", mock.Anything, sessionID).Return(nil)
			},
			expectedError: auth.ErrSessionExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			service := auth.NewService(mockRepo, "test-secret")

			session, err := service.ValidateSession(context.Background(), tt.token)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, session)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, session)
				assert.Equal(t, sessionID, session.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_RefreshSession(t *testing.T) {
	testUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}
	sessionID := uuid.New()

	mockRepo := new(MockUserRepository)
	service := auth.NewService(mockRepo, "test-secret")

	// Generate valid JWT
	validToken, err := service.GenerateJWT(testUser, sessionID)
	require.NoError(t, err)

	validSession := &models.UserSession{
		ID:           sessionID,
		UserID:       testUser.ID,
		SessionToken: validToken,
		ExpiresAt:    time.Now().Add(time.Hour),
		CreatedAt:    time.Now(),
	}

	tests := []struct {
		name          string
		token         string
		mockSetup     func(*MockUserRepository)
		expectedError error
	}{
		{
			name:  "successful refresh",
			token: validToken,
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetSessionByToken", mock.Anything, validToken).Return(validSession, nil)
				repo.On("GetByID", mock.Anything, testUser.ID).Return(testUser, nil)
			},
			expectedError: nil,
		},
		{
			name:  "session not found",
			token: validToken,
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetSessionByToken", mock.Anything, validToken).Return(nil, errors.New("not found"))
			},
			expectedError: auth.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			service := auth.NewService(mockRepo, "test-secret")

			response, err := service.RefreshSession(context.Background(), tt.token)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, testUser.Username, response.User.Username)
				assert.NotEmpty(t, response.SessionToken)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_LogoutUser(t *testing.T) {
	testUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Role:     models.RoleUser,
	}
	sessionID := uuid.New()

	mockRepo := new(MockUserRepository)
	service := auth.NewService(mockRepo, "test-secret")

	// Generate valid JWT
	validToken, err := service.GenerateJWT(testUser, sessionID)
	require.NoError(t, err)

	validSession := &models.UserSession{
		ID:           sessionID,
		UserID:       testUser.ID,
		SessionToken: validToken,
		ExpiresAt:    time.Now().Add(time.Hour),
		CreatedAt:    time.Now(),
	}

	tests := []struct {
		name          string
		token         string
		mockSetup     func(*MockUserRepository)
		expectedError error
	}{
		{
			name:  "successful logout",
			token: validToken,
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetSessionByToken", mock.Anything, validToken).Return(validSession, nil)
				repo.On("DeleteSession", mock.Anything, sessionID).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "invalid session",
			token: "invalid-token",
			mockSetup: func(repo *MockUserRepository) {
				// ValidateSession will fail for invalid token
			},
			expectedError: auth.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			service := auth.NewService(mockRepo, "test-secret")

			err := service.LogoutUser(context.Background(), tt.token)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_GetUserByID(t *testing.T) {
	testUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	tests := []struct {
		name          string
		userID        uuid.UUID
		mockSetup     func(*MockUserRepository)
		expectedUser  *models.User
		expectedError bool
	}{
		{
			name:   "user found",
			userID: testUser.ID,
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetByID", mock.Anything, testUser.ID).Return(testUser, nil)
			},
			expectedUser:  testUser,
			expectedError: false,
		},
		{
			name:   "user not found",
			userID: uuid.New(),
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("not found"))
			},
			expectedUser:  nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			service := auth.NewService(mockRepo, "test-secret")

			user, err := service.GetUserByID(context.Background(), tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser, user)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_GetCurrentUser(t *testing.T) {
	testUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}
	sessionID := uuid.New()

	mockRepo := new(MockUserRepository)
	service := auth.NewService(mockRepo, "test-secret")

	// Generate valid JWT
	validToken, err := service.GenerateJWT(testUser, sessionID)
	require.NoError(t, err)

	validSession := &models.UserSession{
		ID:           sessionID,
		UserID:       testUser.ID,
		SessionToken: validToken,
		ExpiresAt:    time.Now().Add(time.Hour),
		CreatedAt:    time.Now(),
	}

	tests := []struct {
		name          string
		token         string
		mockSetup     func(*MockUserRepository)
		expectedError bool
	}{
		{
			name:  "valid token",
			token: validToken,
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetSessionByToken", mock.Anything, validToken).Return(validSession, nil)
				repo.On("GetByID", mock.Anything, testUser.ID).Return(testUser, nil)
			},
			expectedError: false,
		},
		{
			name:  "invalid token",
			token: "invalid-token",
			mockSetup: func(repo *MockUserRepository) {
				// ValidateSession will fail for invalid token
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			service := auth.NewService(mockRepo, "test-secret")

			user, err := service.GetCurrentUser(context.Background(), tt.token)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testUser.Username, user.Username)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_CleanupExpiredSessions(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*MockUserRepository)
		expectedError bool
	}{
		{
			name: "successful cleanup",
			mockSetup: func(repo *MockUserRepository) {
				repo.On("DeleteExpiredSessions", mock.Anything).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "cleanup error",
			mockSetup: func(repo *MockUserRepository) {
				repo.On("DeleteExpiredSessions", mock.Anything).Return(errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			service := auth.NewService(mockRepo, "test-secret")

			err := service.CleanupExpiredSessions(context.Background())

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
