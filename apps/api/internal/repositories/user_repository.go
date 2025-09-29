package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*models.User, error)
	ListWithFilters(ctx context.Context, filters models.UserListFilters) ([]*models.User, int, error)

	// Session management
	CreateSession(ctx context.Context, session *models.UserSession) error
	GetSessionByToken(ctx context.Context, token string) (*models.UserSession, error)
	DeleteSession(ctx context.Context, sessionID uuid.UUID) error
	DeleteExpiredSessions(ctx context.Context) error
	GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*models.UserSession, error)
}

// userRepository implements UserRepository interface
type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user with password hashing
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	user.UpdatedAt = time.Now()

	query := `
		INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at, last_login, is_active, failed_login_attempts, account_locked_until, password_changed_at, must_change_password)
		VALUES (:id, :username, :email, :password_hash, :role, :created_at, :updated_at, :last_login, :is_active, :failed_login_attempts, :account_locked_until, :password_changed_at, :must_change_password)`

	_, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at, last_login, is_active, failed_login_attempts, account_locked_until, password_changed_at, must_change_password
		FROM users
		WHERE id = $1`

	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// GetByUsername retrieves a user by username (for authentication)
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at, last_login, is_active, failed_login_attempts, account_locked_until, password_changed_at, must_change_password
		FROM users
		WHERE username = $1`

	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at, last_login, is_active, failed_login_attempts, account_locked_until, password_changed_at, must_change_password
		FROM users
		WHERE email = $1`

	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET username = :username, email = :email, password_hash = :password_hash,
		    role = :role, is_active = :is_active, failed_login_attempts = :failed_login_attempts,
		    account_locked_until = :account_locked_until, must_change_password = :must_change_password,
		    password_changed_at = :password_changed_at, updated_at = :updated_at
		WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Delete soft deletes a user (we keep audit trail integrity)
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// For compliance, we don't actually delete users, just deactivate them
	// This preserves audit trail integrity
	query := `
		UPDATE users
		SET role = 'deleted', updated_at = $1
		WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// List retrieves users with pagination (legacy helper)
func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	users, _, err := r.ListWithFilters(ctx, models.UserListFilters{
		Limit:  limit,
		Offset: offset,
	})
	return users, err
}

// ListWithFilters retrieves users with optional filtering and total count
func (r *userRepository) ListWithFilters(ctx context.Context, filters models.UserListFilters) ([]*models.User, int, error) {
	limit := filters.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filters.Offset

	conditions := []string{"role != 'deleted'"}
	args := []interface{}{}

	if filters.Role != "" {
		conditions = append(conditions, fmt.Sprintf("role = $%d", len(args)+1))
		args = append(args, filters.Role)
	}

	if filters.Status != "" {
		switch filters.Status {
		case "active":
			conditions = append(conditions, "is_active = true")
		case "inactive":
			conditions = append(conditions, "is_active = false")
		}
	}

	if filters.Search != "" {
		idx := len(args) + 1
		conditions = append(conditions, fmt.Sprintf("(username ILIKE $%d OR email ILIKE $%d)", idx, idx))
		args = append(args, "%"+filters.Search+"%")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereClause)
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	argsWithPaging := append([]interface{}{}, args...)
	argsWithPaging = append(argsWithPaging, limit, offset)

	query := fmt.Sprintf(`SELECT id, username, email, password_hash, role, created_at, updated_at, last_login, is_active, failed_login_attempts, account_locked_until, password_changed_at, must_change_password
		FROM users
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, len(args)+1, len(args)+2)

	var users []*models.User
	if err := r.db.SelectContext(ctx, &users, query, argsWithPaging...); err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// CreateSession creates a new user session with security tracking
func (r *userRepository) CreateSession(ctx context.Context, session *models.UserSession) error {
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}

	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO user_sessions (id, user_id, session_token, expires_at, created_at, ip_address, user_agent)
		VALUES (:id, :user_id, :session_token, :expires_at, :created_at, :ip_address, :user_agent)`

	_, err := r.db.NamedExecContext(ctx, query, session)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetSessionByToken retrieves a session by token
func (r *userRepository) GetSessionByToken(ctx context.Context, token string) (*models.UserSession, error) {
	var session models.UserSession
	query := `
		SELECT id, user_id, session_token, expires_at, created_at, ip_address, user_agent
		FROM user_sessions
		WHERE session_token = $1`

	err := r.db.GetContext(ctx, &session, query, token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Session not found
		}
		return nil, fmt.Errorf("failed to get session by token: %w", err)
	}

	return &session, nil
}

// DeleteSession deletes a session (logout)
func (r *userRepository) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	query := `DELETE FROM user_sessions WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// DeleteExpiredSessions cleans up expired sessions
func (r *userRepository) DeleteExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM user_sessions WHERE expires_at < $1`

	_, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return nil
}

// GetUserSessions retrieves all active sessions for a user
func (r *userRepository) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*models.UserSession, error) {
	query := `
		SELECT id, user_id, session_token, expires_at, created_at, ip_address, user_agent
		FROM user_sessions
		WHERE user_id = $1 AND expires_at > $2
		ORDER BY created_at DESC`

	var sessions []*models.UserSession
	err := r.db.SelectContext(ctx, &sessions, query, userID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	return sessions, nil
}
