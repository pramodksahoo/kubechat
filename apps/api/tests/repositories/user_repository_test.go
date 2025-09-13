package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/repositories"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	container testcontainers.Container
	db        *sqlx.DB
	repo      repositories.UserRepository
	ctx       context.Context
}

func (suite *UserRepositoryTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Start PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(suite.T(), err)
	suite.container = container

	// Get container connection details
	host, err := container.Host(suite.ctx)
	require.NoError(suite.T(), err)
	port, err := container.MappedPort(suite.ctx, "5432")
	require.NoError(suite.T(), err)

	// Connect to database
	dsn := "host=" + host + " port=" + port.Port() + " user=testuser password=testpass dbname=testdb sslmode=disable"
	db, err := sqlx.Connect("postgres", dsn)
	require.NoError(suite.T(), err)
	suite.db = db

	// Create the schema
	suite.setupSchema()

	// Initialize repository
	suite.repo = repositories.NewUserRepository(db)
}

func (suite *UserRepositoryTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
	if suite.container != nil {
		suite.container.Terminate(suite.ctx)
	}
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	// Clean up data before each test
	suite.cleanupTestData()
}

func (suite *UserRepositoryTestSuite) setupSchema() {
	// Create extensions
	_, err := suite.db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	require.NoError(suite.T(), err)

	// Create users table
	_, err = suite.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username VARCHAR(255) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'user',
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),
			
			CONSTRAINT valid_email CHECK (email ~* '^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,4}$'),
			CONSTRAINT valid_username CHECK (LENGTH(username) >= 3 AND LENGTH(username) <= 255),
			CONSTRAINT valid_role CHECK (role IN ('admin', 'user', 'viewer', 'deleted'))
		);
	`)
	require.NoError(suite.T(), err)

	// Create user_sessions table
	_, err = suite.db.Exec(`
		CREATE TABLE IF NOT EXISTS user_sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			session_token VARCHAR(255) UNIQUE NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			ip_address INET,
			user_agent TEXT,
			
			CONSTRAINT valid_session_token CHECK (LENGTH(session_token) >= 32),
			CONSTRAINT valid_expiry CHECK (expires_at > created_at)
		);
	`)
	require.NoError(suite.T(), err)

	// Create indexes
	suite.db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);`)
	suite.db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`)
	suite.db.Exec(`CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(session_token);`)
	suite.db.Exec(`CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);`)
}

func (suite *UserRepositoryTestSuite) cleanupTestData() {
	suite.db.Exec(`DELETE FROM user_sessions;`)
	suite.db.Exec(`DELETE FROM users;`)
}

func (suite *UserRepositoryTestSuite) TestCreate() {
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}

	err := suite.repo.Create(suite.ctx, user)
	assert.NoError(suite.T(), err)
	assert.NotEqual(suite.T(), uuid.Nil, user.ID)
	assert.False(suite.T(), user.CreatedAt.IsZero())
	assert.False(suite.T(), user.UpdatedAt.IsZero())

	// Verify user exists in database
	var count int
	err = suite.db.Get(&count, "SELECT COUNT(*) FROM users WHERE username = $1", user.Username)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, count)
}

func (suite *UserRepositoryTestSuite) TestCreate_DuplicateUsername() {
	user1 := &models.User{
		Username:     "testuser",
		Email:        "test1@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}

	user2 := &models.User{
		Username:     "testuser", // Same username
		Email:        "test2@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}

	err := suite.repo.Create(suite.ctx, user1)
	require.NoError(suite.T(), err)

	err = suite.repo.Create(suite.ctx, user2)
	assert.Error(suite.T(), err) // Should fail due to unique constraint
}

func (suite *UserRepositoryTestSuite) TestCreate_DuplicateEmail() {
	user1 := &models.User{
		Username:     "testuser1",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}

	user2 := &models.User{
		Username:     "testuser2",
		Email:        "test@example.com", // Same email
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}

	err := suite.repo.Create(suite.ctx, user1)
	require.NoError(suite.T(), err)

	err = suite.repo.Create(suite.ctx, user2)
	assert.Error(suite.T(), err) // Should fail due to unique constraint
}

func (suite *UserRepositoryTestSuite) TestGetByID() {
	// Create a user first
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}
	err := suite.repo.Create(suite.ctx, user)
	require.NoError(suite.T(), err)

	// Retrieve the user
	retrieved, err := suite.repo.GetByID(suite.ctx, user.ID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), retrieved)

	assert.Equal(suite.T(), user.ID, retrieved.ID)
	assert.Equal(suite.T(), user.Username, retrieved.Username)
	assert.Equal(suite.T(), user.Email, retrieved.Email)
	assert.Equal(suite.T(), user.PasswordHash, retrieved.PasswordHash)
	assert.Equal(suite.T(), user.Role, retrieved.Role)
}

func (suite *UserRepositoryTestSuite) TestGetByID_NotFound() {
	randomID := uuid.New()
	user, err := suite.repo.GetByID(suite.ctx, randomID)
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), user)
}

func (suite *UserRepositoryTestSuite) TestGetByUsername() {
	// Create a user first
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}
	err := suite.repo.Create(suite.ctx, user)
	require.NoError(suite.T(), err)

	// Retrieve by username
	retrieved, err := suite.repo.GetByUsername(suite.ctx, user.Username)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), retrieved)

	assert.Equal(suite.T(), user.ID, retrieved.ID)
	assert.Equal(suite.T(), user.Username, retrieved.Username)
	assert.Equal(suite.T(), user.Email, retrieved.Email)
}

func (suite *UserRepositoryTestSuite) TestGetByUsername_NotFound() {
	user, err := suite.repo.GetByUsername(suite.ctx, "nonexistent")
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), user)
}

func (suite *UserRepositoryTestSuite) TestGetByEmail() {
	// Create a user first
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}
	err := suite.repo.Create(suite.ctx, user)
	require.NoError(suite.T(), err)

	// Retrieve by email
	retrieved, err := suite.repo.GetByEmail(suite.ctx, user.Email)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), retrieved)

	assert.Equal(suite.T(), user.ID, retrieved.ID)
	assert.Equal(suite.T(), user.Username, retrieved.Username)
	assert.Equal(suite.T(), user.Email, retrieved.Email)
}

func (suite *UserRepositoryTestSuite) TestGetByEmail_NotFound() {
	user, err := suite.repo.GetByEmail(suite.ctx, "nonexistent@example.com")
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), user)
}

func (suite *UserRepositoryTestSuite) TestUpdate() {
	// Create a user first
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}
	err := suite.repo.Create(suite.ctx, user)
	require.NoError(suite.T(), err)

	originalUpdatedAt := user.UpdatedAt
	time.Sleep(time.Millisecond) // Ensure time difference

	// Update the user
	user.Username = "updateduser"
	user.Email = "updated@example.com"
	user.Role = models.RoleAdmin

	err = suite.repo.Update(suite.ctx, user)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), user.UpdatedAt.After(originalUpdatedAt))

	// Verify the update
	retrieved, err := suite.repo.GetByID(suite.ctx, user.ID)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "updateduser", retrieved.Username)
	assert.Equal(suite.T(), "updated@example.com", retrieved.Email)
	assert.Equal(suite.T(), models.RoleAdmin, retrieved.Role)
}

func (suite *UserRepositoryTestSuite) TestUpdate_NotFound() {
	user := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}

	err := suite.repo.Update(suite.ctx, user)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "user not found")
}

func (suite *UserRepositoryTestSuite) TestDelete() {
	// Create a user first
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}
	err := suite.repo.Create(suite.ctx, user)
	require.NoError(suite.T(), err)

	// Delete the user (soft delete)
	err = suite.repo.Delete(suite.ctx, user.ID)
	require.NoError(suite.T(), err)

	// Verify user is marked as deleted
	var role string
	err = suite.db.Get(&role, "SELECT role FROM users WHERE id = $1", user.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "deleted", role)

	// Verify user doesn't appear in normal queries
	users, err := suite.repo.List(suite.ctx, 10, 0)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), users, 0)
}

func (suite *UserRepositoryTestSuite) TestDelete_NotFound() {
	randomID := uuid.New()
	err := suite.repo.Delete(suite.ctx, randomID)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "user not found")
}

func (suite *UserRepositoryTestSuite) TestList() {
	// Create multiple users
	users := []*models.User{
		{Username: "user1", Email: "user1@example.com", PasswordHash: "hash", Role: models.RoleUser},
		{Username: "user2", Email: "user2@example.com", PasswordHash: "hash", Role: models.RoleAdmin},
		{Username: "user3", Email: "user3@example.com", PasswordHash: "hash", Role: models.RoleViewer},
	}

	for _, user := range users {
		err := suite.repo.Create(suite.ctx, user)
		require.NoError(suite.T(), err)
	}

	// Test list all
	retrieved, err := suite.repo.List(suite.ctx, 10, 0)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), retrieved, 3)

	// Test pagination
	retrieved, err = suite.repo.List(suite.ctx, 2, 0)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), retrieved, 2)

	retrieved, err = suite.repo.List(suite.ctx, 2, 2)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), retrieved, 1)
}

func (suite *UserRepositoryTestSuite) TestList_ExcludesDeleted() {
	// Create users
	user1 := &models.User{Username: "user1", Email: "user1@example.com", PasswordHash: "hash", Role: models.RoleUser}
	user2 := &models.User{Username: "user2", Email: "user2@example.com", PasswordHash: "hash", Role: models.RoleUser}

	err := suite.repo.Create(suite.ctx, user1)
	require.NoError(suite.T(), err)
	err = suite.repo.Create(suite.ctx, user2)
	require.NoError(suite.T(), err)

	// Delete one user
	err = suite.repo.Delete(suite.ctx, user1.ID)
	require.NoError(suite.T(), err)

	// List should only return non-deleted users
	retrieved, err := suite.repo.List(suite.ctx, 10, 0)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), retrieved, 1)
	assert.Equal(suite.T(), user2.ID, retrieved[0].ID)
}

func (suite *UserRepositoryTestSuite) TestCreateSession() {
	// Create a user first
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}
	err := suite.repo.Create(suite.ctx, user)
	require.NoError(suite.T(), err)

	// Create a session
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0"
	session := &models.UserSession{
		UserID:       user.ID,
		SessionToken: "secure_token_12345678901234567890",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		IPAddress:    &ipAddress,
		UserAgent:    &userAgent,
	}

	err = suite.repo.CreateSession(suite.ctx, session)
	require.NoError(suite.T(), err)
	assert.NotEqual(suite.T(), uuid.Nil, session.ID)
	assert.False(suite.T(), session.CreatedAt.IsZero())
}

func (suite *UserRepositoryTestSuite) TestGetSessionByToken() {
	// Create user and session
	user := &models.User{
		Username: "testuser", Email: "test@example.com",
		PasswordHash: "hash", Role: models.RoleUser,
	}
	err := suite.repo.Create(suite.ctx, user)
	require.NoError(suite.T(), err)

	token := "secure_token_12345678901234567890"
	session := &models.UserSession{
		UserID:       user.ID,
		SessionToken: token,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}
	err = suite.repo.CreateSession(suite.ctx, session)
	require.NoError(suite.T(), err)

	// Retrieve session by token
	retrieved, err := suite.repo.GetSessionByToken(suite.ctx, token)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), retrieved)

	assert.Equal(suite.T(), session.ID, retrieved.ID)
	assert.Equal(suite.T(), session.UserID, retrieved.UserID)
	assert.Equal(suite.T(), session.SessionToken, retrieved.SessionToken)
}

func (suite *UserRepositoryTestSuite) TestGetSessionByToken_NotFound() {
	session, err := suite.repo.GetSessionByToken(suite.ctx, "nonexistent_token")
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), session)
}

func (suite *UserRepositoryTestSuite) TestDeleteSession() {
	// Create user and session
	user := &models.User{
		Username: "testuser", Email: "test@example.com",
		PasswordHash: "hash", Role: models.RoleUser,
	}
	err := suite.repo.Create(suite.ctx, user)
	require.NoError(suite.T(), err)

	session := &models.UserSession{
		UserID:       user.ID,
		SessionToken: "secure_token_12345678901234567890",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}
	err = suite.repo.CreateSession(suite.ctx, session)
	require.NoError(suite.T(), err)

	// Delete session
	err = suite.repo.DeleteSession(suite.ctx, session.ID)
	require.NoError(suite.T(), err)

	// Verify session is deleted
	var count int
	err = suite.db.Get(&count, "SELECT COUNT(*) FROM user_sessions WHERE id = $1", session.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, count)
}

func (suite *UserRepositoryTestSuite) TestDeleteExpiredSessions() {
	// Create user
	user := &models.User{
		Username: "testuser", Email: "test@example.com",
		PasswordHash: "hash", Role: models.RoleUser,
	}
	err := suite.repo.Create(suite.ctx, user)
	require.NoError(suite.T(), err)

	// Create expired and valid sessions
	expiredSession := &models.UserSession{
		UserID:       user.ID,
		SessionToken: "expired_token_123456789012345678901234567890",
		ExpiresAt:    time.Now().Add(-time.Hour), // Expired
	}
	validSession := &models.UserSession{
		UserID:       user.ID,
		SessionToken: "valid_token_1234567890123456789012345678901234567890",
		ExpiresAt:    time.Now().Add(time.Hour), // Valid
	}

	err = suite.repo.CreateSession(suite.ctx, expiredSession)
	require.NoError(suite.T(), err)
	err = suite.repo.CreateSession(suite.ctx, validSession)
	require.NoError(suite.T(), err)

	// Delete expired sessions
	err = suite.repo.DeleteExpiredSessions(suite.ctx)
	require.NoError(suite.T(), err)

	// Verify only valid session remains
	sessions, err := suite.repo.GetUserSessions(suite.ctx, user.ID)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), sessions, 1)
	assert.Equal(suite.T(), validSession.ID, sessions[0].ID)
}

func (suite *UserRepositoryTestSuite) TestGetUserSessions() {
	// Create user
	user := &models.User{
		Username: "testuser", Email: "test@example.com",
		PasswordHash: "hash", Role: models.RoleUser,
	}
	err := suite.repo.Create(suite.ctx, user)
	require.NoError(suite.T(), err)

	// Create multiple sessions
	session1 := &models.UserSession{
		UserID:       user.ID,
		SessionToken: "token1_1234567890123456789012345678901234567890",
		ExpiresAt:    time.Now().Add(time.Hour),
	}
	session2 := &models.UserSession{
		UserID:       user.ID,
		SessionToken: "token2_1234567890123456789012345678901234567890",
		ExpiresAt:    time.Now().Add(2 * time.Hour),
	}

	err = suite.repo.CreateSession(suite.ctx, session1)
	require.NoError(suite.T(), err)
	err = suite.repo.CreateSession(suite.ctx, session2)
	require.NoError(suite.T(), err)

	// Retrieve user sessions
	sessions, err := suite.repo.GetUserSessions(suite.ctx, user.ID)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), sessions, 2)

	// Sessions should be ordered by created_at DESC
	assert.True(suite.T(), sessions[0].CreatedAt.After(sessions[1].CreatedAt) ||
		sessions[0].CreatedAt.Equal(sessions[1].CreatedAt))
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
