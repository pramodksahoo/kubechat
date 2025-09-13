package user

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// MockRepository is a mock implementation of the Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) List(ctx context.Context, offset, limit int) ([]*models.User, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockRepository) CreateSession(ctx context.Context, session *models.UserSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockRepository) GetSessionByToken(ctx context.Context, token string) (*models.UserSession, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserSession), args.Error(1)
}

func (m *MockRepository) DeleteSession(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func TestNew(t *testing.T) {
	repo := &MockRepository{}
	service := New(repo)
	assert.NotNil(t, service)
}

func TestService_CreateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		repo := &MockRepository{}
		service := New(repo)

		req := models.CreateUserRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
			Role:     models.RoleUser,
		}

		// Mock repository calls
		repo.On("GetByUsername", ctx, "testuser").Return(nil, errors.New("not found"))
		repo.On("GetByEmail", ctx, "test@example.com").Return(nil, errors.New("not found"))
		repo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(nil)

		user, err := service.CreateUser(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, models.RoleUser, user.Role)
		assert.NotEmpty(t, user.PasswordHash)

		repo.AssertExpectations(t)
	})

	t.Run("invalid role", func(t *testing.T) {
		repo := &MockRepository{}
		service := New(repo)

		req := models.CreateUserRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
			Role:     "invalid",
		}

		user, err := service.CreateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "invalid role")
	})

	t.Run("username already exists", func(t *testing.T) {
		repo := &MockRepository{}
		service := New(repo)

		req := models.CreateUserRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		existingUser := &models.User{Username: "testuser"}
		repo.On("GetByUsername", ctx, "testuser").Return(existingUser, nil)

		user, err := service.CreateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "username already exists")

		repo.AssertExpectations(t)
	})

	t.Run("email already exists", func(t *testing.T) {
		repo := &MockRepository{}
		service := New(repo)

		req := models.CreateUserRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		existingUser := &models.User{Email: "test@example.com"}
		repo.On("GetByUsername", ctx, "testuser").Return(nil, errors.New("not found"))
		repo.On("GetByEmail", ctx, "test@example.com").Return(existingUser, nil)

		user, err := service.CreateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "email already exists")

		repo.AssertExpectations(t)
	})
}

func TestService_GetUserByID(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		repo := &MockRepository{}
		service := New(repo)

		userID := uuid.New()
		expectedUser := &models.User{
			ID:       userID,
			Username: "testuser",
			Email:    "test@example.com",
		}

		repo.On("GetByID", ctx, userID).Return(expectedUser, nil)

		user, err := service.GetUserByID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedUser, user)

		repo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		repo := &MockRepository{}
		service := New(repo)

		userID := uuid.New()
		repo.On("GetByID", ctx, userID).Return(nil, errors.New("not found"))

		user, err := service.GetUserByID(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, user)

		repo.AssertExpectations(t)
	})
}

func TestService_AuthenticateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("successful authentication", func(t *testing.T) {
		repo := &MockRepository{}
		service := New(repo)

		// Create a user with known password
		user := &models.User{
			ID:       uuid.New(),
			Username: "testuser",
			Email:    "test@example.com",
			Role:     models.RoleUser,
		}
		_ = user.HashPassword("password123")

		req := models.LoginRequest{
			Username: "testuser",
			Password: "password123",
		}

		repo.On("GetByUsername", ctx, "testuser").Return(user, nil)
		repo.On("CreateSession", ctx, mock.AnythingOfType("*models.UserSession")).Return(nil)

		response, err := service.AuthenticateUser(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, user.ID, response.User.ID)
		assert.NotEmpty(t, response.SessionToken)

		repo.AssertExpectations(t)
	})

	t.Run("invalid username", func(t *testing.T) {
		repo := &MockRepository{}
		service := New(repo)

		req := models.LoginRequest{
			Username: "nonexistent",
			Password: "password123",
		}

		repo.On("GetByUsername", ctx, "nonexistent").Return(nil, errors.New("not found"))

		response, err := service.AuthenticateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "authentication failed")

		repo.AssertExpectations(t)
	})

	t.Run("invalid password", func(t *testing.T) {
		repo := &MockRepository{}
		service := New(repo)

		user := &models.User{
			ID:       uuid.New(),
			Username: "testuser",
			Email:    "test@example.com",
			Role:     models.RoleUser,
		}
		_ = user.HashPassword("correctpassword")

		req := models.LoginRequest{
			Username: "testuser",
			Password: "wrongpassword",
		}

		repo.On("GetByUsername", ctx, "testuser").Return(user, nil)

		response, err := service.AuthenticateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "authentication failed")

		repo.AssertExpectations(t)
	})
}

func TestService_ListUsers(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list", func(t *testing.T) {
		repo := &MockRepository{}
		service := New(repo)

		expectedUsers := []*models.User{
			{ID: uuid.New(), Username: "user1"},
			{ID: uuid.New(), Username: "user2"},
		}

		repo.On("List", ctx, 0, 50).Return(expectedUsers, nil)

		users, err := service.ListUsers(ctx, 0, 50)
		require.NoError(t, err)
		assert.Equal(t, expectedUsers, users)

		repo.AssertExpectations(t)
	})

	t.Run("with limit normalization", func(t *testing.T) {
		repo := &MockRepository{}
		service := New(repo)

		expectedUsers := []*models.User{}

		// Test limit too high gets normalized
		repo.On("List", ctx, 0, 50).Return(expectedUsers, nil)

		users, err := service.ListUsers(ctx, 0, 200)
		require.NoError(t, err)
		assert.Equal(t, expectedUsers, users)

		repo.AssertExpectations(t)
	})

	t.Run("with offset normalization", func(t *testing.T) {
		repo := &MockRepository{}
		service := New(repo)

		expectedUsers := []*models.User{}

		// Test negative offset gets normalized
		repo.On("List", ctx, 0, 10).Return(expectedUsers, nil)

		users, err := service.ListUsers(ctx, -5, 10)
		require.NoError(t, err)
		assert.Equal(t, expectedUsers, users)

		repo.AssertExpectations(t)
	})
}

func TestService_ValidateSession(t *testing.T) {
	ctx := context.Background()

	t.Run("valid session", func(t *testing.T) {
		repo := &MockRepository{}
		service := New(repo)

		userID := uuid.New()
		user := &models.User{
			ID:       userID,
			Username: "testuser",
		}

		session, _ := models.NewUserSession(userID, nil, nil)

		repo.On("GetSessionByToken", ctx, session.SessionToken).Return(session, nil)
		repo.On("GetByID", ctx, userID).Return(user, nil)

		validatedUser, err := service.ValidateSession(ctx, session.SessionToken)
		require.NoError(t, err)
		assert.Equal(t, user, validatedUser)

		repo.AssertExpectations(t)
	})

	t.Run("session not found", func(t *testing.T) {
		repo := &MockRepository{}
		service := New(repo)

		repo.On("GetSessionByToken", ctx, "invalid-token").Return(nil, errors.New("not found"))

		user, err := service.ValidateSession(ctx, "invalid-token")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "invalid session token")

		repo.AssertExpectations(t)
	})
}
