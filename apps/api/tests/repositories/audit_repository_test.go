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

type AuditRepositoryTestSuite struct {
	suite.Suite
	container testcontainers.Container
	db        *sqlx.DB
	repo      repositories.AuditRepository
	ctx       context.Context
	testUser  *models.User
}

func (suite *AuditRepositoryTestSuite) SetupSuite() {
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
	suite.repo = repositories.NewAuditRepository(db)

	// Create a test user for audit logs
	suite.createTestUser()
}

func (suite *AuditRepositoryTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
	if suite.container != nil {
		suite.container.Terminate(suite.ctx)
	}
}

func (suite *AuditRepositoryTestSuite) SetupTest() {
	// Clean up audit logs before each test
	suite.db.Exec(`DELETE FROM audit_logs;`)
}

func (suite *AuditRepositoryTestSuite) setupSchema() {
	// Create extensions
	_, err := suite.db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	require.NoError(suite.T(), err)
	_, err = suite.db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto";`)
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
			updated_at TIMESTAMP DEFAULT NOW()
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
			created_at TIMESTAMP DEFAULT NOW()
		);
	`)
	require.NoError(suite.T(), err)

	// Create audit_logs table with all integrity features
	_, err = suite.db.Exec(`
		CREATE TABLE IF NOT EXISTS audit_logs (
			id BIGSERIAL PRIMARY KEY,
			user_id UUID REFERENCES users(id),
			session_id UUID REFERENCES user_sessions(id),
			query_text TEXT NOT NULL,
			generated_command TEXT NOT NULL,
			safety_level VARCHAR(20) NOT NULL CHECK (safety_level IN ('safe', 'warning', 'dangerous')),
			execution_result JSONB,
			execution_status VARCHAR(20) NOT NULL CHECK (execution_status IN ('success', 'failed', 'cancelled')),
			cluster_context VARCHAR(255),
			namespace_context VARCHAR(255),
			timestamp TIMESTAMP DEFAULT NOW() NOT NULL,
			ip_address INET,
			user_agent TEXT,
			checksum VARCHAR(64) NOT NULL,
			previous_checksum VARCHAR(64)
		);
	`)
	require.NoError(suite.T(), err)

	// Create checksum calculation function
	_, err = suite.db.Exec(`
		CREATE OR REPLACE FUNCTION calculate_audit_checksum(
			p_user_id UUID,
			p_session_id UUID,
			p_query_text TEXT,
			p_generated_command TEXT,
			p_safety_level VARCHAR(20),
			p_execution_result JSONB,
			p_execution_status VARCHAR(20),
			p_cluster_context VARCHAR(255),
			p_namespace_context VARCHAR(255),
			p_timestamp TIMESTAMP,
			p_ip_address INET,
			p_user_agent TEXT,
			p_previous_checksum VARCHAR(64)
		)
		RETURNS VARCHAR(64) AS $$
		DECLARE
			checksum_input TEXT;
		BEGIN
			checksum_input := COALESCE(p_user_id::TEXT, '') || '|' ||
							  COALESCE(p_session_id::TEXT, '') || '|' ||
							  COALESCE(p_query_text, '') || '|' ||
							  COALESCE(p_generated_command, '') || '|' ||
							  COALESCE(p_safety_level, '') || '|' ||
							  COALESCE(p_execution_result::TEXT, '') || '|' ||
							  COALESCE(p_execution_status, '') || '|' ||
							  COALESCE(p_cluster_context, '') || '|' ||
							  COALESCE(p_namespace_context, '') || '|' ||
							  COALESCE(p_timestamp::TEXT, '') || '|' ||
							  COALESCE(p_ip_address::TEXT, '') || '|' ||
							  COALESCE(p_user_agent, '') || '|' ||
							  COALESCE(p_previous_checksum, '');
			
			RETURN encode(digest(checksum_input, 'sha256'), 'hex');
		END;
		$$ language 'plpgsql';
	`)
	require.NoError(suite.T(), err)

	// Create trigger function for automatic checksum calculation
	_, err = suite.db.Exec(`
		CREATE OR REPLACE FUNCTION set_audit_log_checksum()
		RETURNS TRIGGER AS $$
		DECLARE
			prev_checksum VARCHAR(64);
		BEGIN
			SELECT checksum INTO prev_checksum
			FROM audit_logs
			ORDER BY id DESC
			LIMIT 1;
			
			NEW.checksum := calculate_audit_checksum(
				NEW.user_id,
				NEW.session_id,
				NEW.query_text,
				NEW.generated_command,
				NEW.safety_level,
				NEW.execution_result,
				NEW.execution_status,
				NEW.cluster_context,
				NEW.namespace_context,
				NEW.timestamp,
				NEW.ip_address,
				NEW.user_agent,
				prev_checksum
			);
			
			NEW.previous_checksum := prev_checksum;
			
			RETURN NEW;
		END;
		$$ language 'plpgsql';
	`)
	require.NoError(suite.T(), err)

	// Create integrity verification function
	_, err = suite.db.Exec(`
		CREATE OR REPLACE FUNCTION verify_audit_log_integrity(log_id BIGINT DEFAULT NULL)
		RETURNS TABLE(log_id BIGINT, is_valid BOOLEAN, error_message TEXT) AS $$
		DECLARE
			audit_record RECORD;
			calculated_checksum VARCHAR(64);
			prev_checksum VARCHAR(64);
		BEGIN
			FOR audit_record IN 
				SELECT * FROM audit_logs 
				WHERE (verify_audit_log_integrity.log_id IS NULL OR audit_logs.id = verify_audit_log_integrity.log_id)
				ORDER BY audit_logs.id
			LOOP
				SELECT audit_logs.checksum INTO prev_checksum
				FROM audit_logs
				WHERE audit_logs.id < audit_record.id
				ORDER BY audit_logs.id DESC
				LIMIT 1;
				
				calculated_checksum := calculate_audit_checksum(
					audit_record.user_id,
					audit_record.session_id,
					audit_record.query_text,
					audit_record.generated_command,
					audit_record.safety_level,
					audit_record.execution_result,
					audit_record.execution_status,
					audit_record.cluster_context,
					audit_record.namespace_context,
					audit_record.timestamp,
					audit_record.ip_address,
					audit_record.user_agent,
					prev_checksum
				);
				
				IF audit_record.checksum = calculated_checksum AND 
				   (audit_record.previous_checksum = prev_checksum OR 
					(audit_record.previous_checksum IS NULL AND prev_checksum IS NULL)) THEN
					RETURN QUERY SELECT audit_record.id, true, NULL::TEXT;
				ELSE
					RETURN QUERY SELECT audit_record.id, false, 
						'Checksum mismatch: expected ' || calculated_checksum || 
						', got ' || audit_record.checksum;
				END IF;
			END LOOP;
		END;
		$$ language 'plpgsql';
	`)
	require.NoError(suite.T(), err)

	// Create trigger
	_, err = suite.db.Exec(`
		DROP TRIGGER IF EXISTS set_audit_log_checksum_trigger ON audit_logs;
		CREATE TRIGGER set_audit_log_checksum_trigger BEFORE INSERT ON audit_logs
			FOR EACH ROW EXECUTE FUNCTION set_audit_log_checksum();
	`)
	require.NoError(suite.T(), err)

	// Create indexes
	suite.db.Exec(`CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);`)
	suite.db.Exec(`CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp);`)
	suite.db.Exec(`CREATE INDEX IF NOT EXISTS idx_audit_logs_checksum ON audit_logs(checksum);`)
}

func (suite *AuditRepositoryTestSuite) createTestUser() {
	suite.testUser = &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleUser,
	}

	_, err := suite.db.NamedExec(`
		INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at)
		VALUES (:id, :username, :email, :password_hash, :role, :created_at, :updated_at)
	`, suite.testUser)
	require.NoError(suite.T(), err)
}

func (suite *AuditRepositoryTestSuite) TestCreate() {
	executionResult := map[string]interface{}{
		"stdout":    "pod1\npod2\npod3",
		"stderr":    "",
		"exit_code": 0,
	}

	audit := &models.AuditLog{
		UserID:           &suite.testUser.ID,
		QueryText:        "show me pods",
		GeneratedCommand: "kubectl get pods",
		SafetyLevel:      "safe",
		ExecutionResult:  executionResult,
		ExecutionStatus:  "success",
		ClusterContext:   &[]string{"production"}[0],
		NamespaceContext: &[]string{"default"}[0],
		Timestamp:        time.Now(),
		IPAddress:        &[]string{"192.168.1.1"}[0],
		UserAgent:        &[]string{"Mozilla/5.0"}[0],
	}

	err := suite.repo.CreateAuditLog(suite.ctx, audit)
	require.NoError(suite.T(), err)

	// Verify audit log was created with proper checksum
	assert.NotZero(suite.T(), audit.ID)
	assert.NotEmpty(suite.T(), audit.Checksum)
	assert.Len(suite.T(), audit.Checksum, 64) // SHA-256 hex length

	// Verify in database
	var count int
	err = suite.db.Get(&count, "SELECT COUNT(*) FROM audit_logs WHERE id = $1", audit.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, count)
}

func (suite *AuditRepositoryTestSuite) TestCreate_ChecksumChaining() {
	// Create first audit log
	audit1 := &models.AuditLog{
		UserID:           &suite.testUser.ID,
		QueryText:        "first query",
		GeneratedCommand: "kubectl get pods",
		SafetyLevel:      "safe",
		ExecutionResult:  map[string]interface{}{"result": "success"},
		ExecutionStatus:  "success",
		Timestamp:        time.Now(),
	}

	err := suite.repo.CreateAuditLog(suite.ctx, audit1)
	require.NoError(suite.T(), err)

	// Create second audit log
	audit2 := &models.AuditLog{
		UserID:           &suite.testUser.ID,
		QueryText:        "second query",
		GeneratedCommand: "kubectl get services",
		SafetyLevel:      "safe",
		ExecutionResult:  map[string]interface{}{"result": "success"},
		ExecutionStatus:  "success",
		Timestamp:        time.Now(),
	}

	err = suite.repo.CreateAuditLog(suite.ctx, audit2)
	require.NoError(suite.T(), err)

	// Verify checksum chaining
	assert.Nil(suite.T(), audit1.PreviousChecksum) // First entry has no previous
	assert.NotNil(suite.T(), audit2.PreviousChecksum)
	if audit2.PreviousChecksum != nil {
		assert.Equal(suite.T(), audit1.Checksum, *audit2.PreviousChecksum)
	}
}

func (suite *AuditRepositoryTestSuite) TestGetByID() {
	// Create an audit log first
	audit := &models.AuditLog{
		UserID:           &suite.testUser.ID,
		QueryText:        "show me pods",
		GeneratedCommand: "kubectl get pods",
		SafetyLevel:      "safe",
		ExecutionResult:  map[string]interface{}{"test": "data"},
		ExecutionStatus:  "success",
		Timestamp:        time.Now(),
	}

	err := suite.repo.CreateAuditLog(suite.ctx, audit)
	require.NoError(suite.T(), err)

	// Retrieve by ID
	retrieved, err := suite.repo.GetAuditLogByID(suite.ctx, audit.ID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), retrieved)

	assert.Equal(suite.T(), audit.ID, retrieved.ID)
	assert.Equal(suite.T(), audit.QueryText, retrieved.QueryText)
	assert.Equal(suite.T(), audit.GeneratedCommand, retrieved.GeneratedCommand)
	assert.Equal(suite.T(), audit.SafetyLevel, retrieved.SafetyLevel)
	assert.Equal(suite.T(), audit.ExecutionStatus, retrieved.ExecutionStatus)
}

func (suite *AuditRepositoryTestSuite) TestGetByID_NotFound() {
	retrieved, err := suite.repo.GetAuditLogByID(suite.ctx, 99999)
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), retrieved)
}

func (suite *AuditRepositoryTestSuite) TestList() {
	// Create multiple audit logs
	audits := []*models.AuditLog{
		{UserID: &suite.testUser.ID, QueryText: "query1", GeneratedCommand: "cmd1", SafetyLevel: "safe", ExecutionStatus: "success", ExecutionResult: map[string]interface{}{}},
		{UserID: &suite.testUser.ID, QueryText: "query2", GeneratedCommand: "cmd2", SafetyLevel: "warning", ExecutionStatus: "failed", ExecutionResult: map[string]interface{}{}},
		{UserID: &suite.testUser.ID, QueryText: "query3", GeneratedCommand: "cmd3", SafetyLevel: "dangerous", ExecutionStatus: "success", ExecutionResult: map[string]interface{}{}},
	}

	for _, audit := range audits {
		err := suite.repo.CreateAuditLog(suite.ctx, audit)
		require.NoError(suite.T(), err)
		time.Sleep(time.Millisecond) // Ensure different timestamps
	}

	// Test list all
	filter := models.AuditLogFilter{Limit: 10}
	retrieved, err := suite.repo.GetAuditLogs(suite.ctx, filter)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), retrieved, 3)

	// Should be ordered by timestamp DESC
	assert.True(suite.T(), retrieved[0].Timestamp.After(retrieved[1].Timestamp))
	assert.True(suite.T(), retrieved[1].Timestamp.After(retrieved[2].Timestamp))
}

func (suite *AuditRepositoryTestSuite) TestList_WithFilters() {
	// Create audit logs with different properties
	userID2 := uuid.New()
	_, err := suite.db.Exec("INSERT INTO users (id, username, email, password_hash, role) VALUES ($1, 'user2', 'user2@test.com', 'hash', 'user')", userID2)
	require.NoError(suite.T(), err)

	audits := []*models.AuditLog{
		{UserID: &suite.testUser.ID, QueryText: "query1", GeneratedCommand: "cmd1", SafetyLevel: "safe", ExecutionStatus: "success", ExecutionResult: map[string]interface{}{}},
		{UserID: &userID2, QueryText: "query2", GeneratedCommand: "cmd2", SafetyLevel: "warning", ExecutionStatus: "failed", ExecutionResult: map[string]interface{}{}},
		{UserID: &suite.testUser.ID, QueryText: "query3", GeneratedCommand: "cmd3", SafetyLevel: "dangerous", ExecutionStatus: "success", ExecutionResult: map[string]interface{}{}},
	}

	for _, audit := range audits {
		err := suite.repo.CreateAuditLog(suite.ctx, audit)
		require.NoError(suite.T(), err)
	}

	// Filter by user ID
	filter := models.AuditLogFilter{UserID: &suite.testUser.ID, Limit: 10}
	retrieved, err := suite.repo.GetAuditLogs(suite.ctx, filter)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), retrieved, 2)

	// Filter by safety level
	safetyLevel := "warning"
	filter = models.AuditLogFilter{SafetyLevel: &safetyLevel, Limit: 10}
	retrieved, err = suite.repo.GetAuditLogs(suite.ctx, filter)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), retrieved, 1)
	assert.Equal(suite.T(), "warning", retrieved[0].SafetyLevel)

	// Filter by status
	status := "success"
	filter = models.AuditLogFilter{Status: &status, Limit: 10}
	retrieved, err = suite.repo.GetAuditLogs(suite.ctx, filter)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), retrieved, 2)
}

func (suite *AuditRepositoryTestSuite) TestGetByUser() {
	// Create audit logs for different users
	userID2 := uuid.New()
	_, err := suite.db.Exec("INSERT INTO users (id, username, email, password_hash, role) VALUES ($1, 'user2', 'user2@test.com', 'hash', 'user')", userID2)
	require.NoError(suite.T(), err)

	// Create audit logs
	audits := []*models.AuditLog{
		{UserID: &suite.testUser.ID, QueryText: "query1", GeneratedCommand: "cmd1", SafetyLevel: "safe", ExecutionStatus: "success", ExecutionResult: map[string]interface{}{}},
		{UserID: &userID2, QueryText: "query2", GeneratedCommand: "cmd2", SafetyLevel: "safe", ExecutionStatus: "success", ExecutionResult: map[string]interface{}{}},
		{UserID: &suite.testUser.ID, QueryText: "query3", GeneratedCommand: "cmd3", SafetyLevel: "safe", ExecutionStatus: "success", ExecutionResult: map[string]interface{}{}},
	}

	for _, audit := range audits {
		err := suite.repo.CreateAuditLog(suite.ctx, audit)
		require.NoError(suite.T(), err)
	}

	// Get audits for test user
	retrieved, err := suite.repo.GetAuditLogsByUserID(suite.ctx, suite.testUser.ID.String(), 10, 0)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), retrieved, 2)

	for _, audit := range retrieved {
		assert.Equal(suite.T(), suite.testUser.ID, *audit.UserID)
	}
}

func (suite *AuditRepositoryTestSuite) TestVerifyIntegrity() {
	// Create multiple audit logs to test chain integrity
	audits := []*models.AuditLog{
		{UserID: &suite.testUser.ID, QueryText: "query1", GeneratedCommand: "cmd1", SafetyLevel: "safe", ExecutionStatus: "success", ExecutionResult: map[string]interface{}{}},
		{UserID: &suite.testUser.ID, QueryText: "query2", GeneratedCommand: "cmd2", SafetyLevel: "safe", ExecutionStatus: "success", ExecutionResult: map[string]interface{}{}},
		{UserID: &suite.testUser.ID, QueryText: "query3", GeneratedCommand: "cmd3", SafetyLevel: "safe", ExecutionStatus: "success", ExecutionResult: map[string]interface{}{}},
	}

	for _, audit := range audits {
		err := suite.repo.CreateAuditLog(suite.ctx, audit)
		require.NoError(suite.T(), err)
	}

	// Verify integrity of all logs
	results, err := suite.repo.VerifyIntegrity(suite.ctx, 1, 1000)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 3)

	// All should be valid
	for _, result := range results {
		assert.True(suite.T(), result.IsValid, "Audit log %d should be valid: %s", result.LogID, result.ErrorMessage)
	}
}

func (suite *AuditRepositoryTestSuite) TestVerifyIntegrity_SingleLog() {
	// Create an audit log
	audit := &models.AuditLog{
		UserID:           &suite.testUser.ID,
		QueryText:        "test query",
		GeneratedCommand: "test command",
		SafetyLevel:      "safe",
		ExecutionStatus:  "success",
		ExecutionResult:  map[string]interface{}{},
	}

	err := suite.repo.CreateAuditLog(suite.ctx, audit)
	require.NoError(suite.T(), err)

	// Verify integrity of specific log
	results, err := suite.repo.VerifyIntegrity(suite.ctx, audit.ID, audit.ID)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 1)
	assert.Equal(suite.T(), audit.ID, results[0].LogID)
	assert.True(suite.T(), results[0].IsValid)
}

func (suite *AuditRepositoryTestSuite) TestVerifyIntegrity_TamperedData() {
	// Create an audit log
	audit := &models.AuditLog{
		UserID:           &suite.testUser.ID,
		QueryText:        "test query",
		GeneratedCommand: "test command",
		SafetyLevel:      "safe",
		ExecutionStatus:  "success",
		ExecutionResult:  map[string]interface{}{},
	}

	err := suite.repo.CreateAuditLog(suite.ctx, audit)
	require.NoError(suite.T(), err)

	// Tamper with the data (bypass triggers by direct update)
	_, err = suite.db.Exec(`
		UPDATE audit_logs 
		SET query_text = 'tampered query' 
		WHERE id = $1
	`, audit.ID)
	require.NoError(suite.T(), err)

	// Verify integrity - should detect tampering
	results, err := suite.repo.VerifyIntegrity(suite.ctx, audit.ID, audit.ID)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 1)
	assert.False(suite.T(), results[0].IsValid)
	assert.Contains(suite.T(), results[0].ErrorMessage, "Checksum mismatch")
}

func (suite *AuditRepositoryTestSuite) TestGetLastChecksum() {
	// Initially should return nil
	checksum, err := suite.repo.GetLastChecksum(suite.ctx)
	require.NoError(suite.T(), err)
	assert.Nil(suite.T(), checksum)

	// Create an audit log
	audit := &models.AuditLog{
		UserID:           &suite.testUser.ID,
		QueryText:        "test query",
		GeneratedCommand: "test command",
		SafetyLevel:      "safe",
		ExecutionStatus:  "success",
		ExecutionResult:  map[string]interface{}{},
	}

	err = suite.repo.CreateAuditLog(suite.ctx, audit)
	require.NoError(suite.T(), err)

	// Now should return the checksum
	checksum, err = suite.repo.GetLastChecksum(suite.ctx)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), checksum)
	assert.Equal(suite.T(), audit.Checksum, *checksum)
}

func (suite *AuditRepositoryTestSuite) TestGetSummary() {
	// Create audit logs with different properties
	audits := []*models.AuditLog{
		{UserID: &suite.testUser.ID, QueryText: "query1", GeneratedCommand: "cmd1", SafetyLevel: "safe", ExecutionStatus: "success", ExecutionResult: map[string]interface{}{}},
		{UserID: &suite.testUser.ID, QueryText: "query2", GeneratedCommand: "cmd2", SafetyLevel: "warning", ExecutionStatus: "failed", ExecutionResult: map[string]interface{}{}},
		{UserID: &suite.testUser.ID, QueryText: "query3", GeneratedCommand: "cmd3", SafetyLevel: "dangerous", ExecutionStatus: "success", ExecutionResult: map[string]interface{}{}},
		{UserID: &suite.testUser.ID, QueryText: "query4", GeneratedCommand: "cmd4", SafetyLevel: "safe", ExecutionStatus: "cancelled", ExecutionResult: map[string]interface{}{}},
	}

	for _, audit := range audits {
		err := suite.repo.CreateAuditLog(suite.ctx, audit)
		require.NoError(suite.T(), err)
	}

	// Get summary for all logs
	summary, err := suite.repo.GetAuditLogSummary(suite.ctx, models.AuditLogFilter{})
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), summary)

	assert.Equal(suite.T(), 4, summary.TotalEntries)
	assert.Equal(suite.T(), 2, summary.SafeOperations)
	assert.Equal(suite.T(), 1, summary.WarningOps)
	assert.Equal(suite.T(), 1, summary.DangerousOps)
	assert.Equal(suite.T(), 2, summary.SuccessfulOps)
	assert.Equal(suite.T(), 1, summary.FailedOps)
	assert.Equal(suite.T(), 1, summary.CancelledOps)
}

func (suite *AuditRepositoryTestSuite) TestGetSummary_WithUserFilter() {
	// Create another user
	userID2 := uuid.New()
	_, err := suite.db.Exec("INSERT INTO users (id, username, email, password_hash, role) VALUES ($1, 'user2', 'user2@test.com', 'hash', 'user')", userID2)
	require.NoError(suite.T(), err)

	// Create audit logs for both users
	audits := []*models.AuditLog{
		{UserID: &suite.testUser.ID, QueryText: "query1", GeneratedCommand: "cmd1", SafetyLevel: "safe", ExecutionStatus: "success", ExecutionResult: map[string]interface{}{}},
		{UserID: &userID2, QueryText: "query2", GeneratedCommand: "cmd2", SafetyLevel: "safe", ExecutionStatus: "success", ExecutionResult: map[string]interface{}{}},
		{UserID: &suite.testUser.ID, QueryText: "query3", GeneratedCommand: "cmd3", SafetyLevel: "safe", ExecutionStatus: "success", ExecutionResult: map[string]interface{}{}},
	}

	for _, audit := range audits {
		err := suite.repo.CreateAuditLog(suite.ctx, audit)
		require.NoError(suite.T(), err)
	}

	// Get summary filtered by test user
	summary, err := suite.repo.GetAuditLogSummary(suite.ctx, models.AuditLogFilter{UserID: &suite.testUser.ID})
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), 2, summary.TotalEntries) // Only 2 for testUser
	assert.Equal(suite.T(), 2, summary.SafeOperations)
	assert.Equal(suite.T(), 2, summary.SuccessfulOps)
}

func (suite *AuditRepositoryTestSuite) TestGetDangerousOperations() {
	// Create audit logs with different safety levels
	audits := []*models.AuditLog{
		{UserID: &suite.testUser.ID, QueryText: "query1", GeneratedCommand: "cmd1", SafetyLevel: "safe", ExecutionStatus: "success", ExecutionResult: map[string]interface{}{}},
		{UserID: &suite.testUser.ID, QueryText: "query2", GeneratedCommand: "cmd2", SafetyLevel: "dangerous", ExecutionStatus: "success", ExecutionResult: map[string]interface{}{}},
		{UserID: &suite.testUser.ID, QueryText: "query3", GeneratedCommand: "cmd3", SafetyLevel: "warning", ExecutionStatus: "success", ExecutionResult: map[string]interface{}{}},
		{UserID: &suite.testUser.ID, QueryText: "query4", GeneratedCommand: "cmd4", SafetyLevel: "dangerous", ExecutionStatus: "success", ExecutionResult: map[string]interface{}{}},
	}

	for _, audit := range audits {
		err := suite.repo.CreateAuditLog(suite.ctx, audit)
		require.NoError(suite.T(), err)
	}

	// Get dangerous operations
	dangerous, err := suite.repo.GetDangerousOperations(suite.ctx, models.AuditLogFilter{Limit: 10})
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), dangerous, 2)

	// All returned should be dangerous
	for _, audit := range dangerous {
		assert.Equal(suite.T(), "dangerous", audit.SafetyLevel)
	}
}

func (suite *AuditRepositoryTestSuite) TestConcurrentAuditCreation() {
	// Test that concurrent audit log creation maintains integrity
	const numGoroutines = 10
	const logsPerGoroutine = 5

	type result struct {
		err   error
		audit *models.AuditLog
	}

	results := make(chan result, numGoroutines*logsPerGoroutine)

	// Create audit logs concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			for j := 0; j < logsPerGoroutine; j++ {
				audit := &models.AuditLog{
					UserID:           &suite.testUser.ID,
					QueryText:        "concurrent query",
					GeneratedCommand: "concurrent command",
					SafetyLevel:      "safe",
					ExecutionStatus:  "success",
					ExecutionResult:  map[string]interface{}{"routine": routineID, "log": j},
				}

				err := suite.repo.CreateAuditLog(suite.ctx, audit)
				results <- result{err: err, audit: audit}
			}
		}(i)
	}

	// Collect results
	var audits []*models.AuditLog
	for i := 0; i < numGoroutines*logsPerGoroutine; i++ {
		res := <-results
		require.NoError(suite.T(), res.err)
		audits = append(audits, res.audit)
	}

	// Verify all have unique IDs and checksums
	ids := make(map[int64]bool)
	checksums := make(map[string]bool)

	for _, audit := range audits {
		assert.False(suite.T(), ids[audit.ID], "Duplicate ID found: %d", audit.ID)
		ids[audit.ID] = true

		assert.NotEmpty(suite.T(), audit.Checksum)
		assert.False(suite.T(), checksums[audit.Checksum], "Duplicate checksum found: %s", audit.Checksum)
		checksums[audit.Checksum] = true
	}

	// Verify integrity of all created logs
	results_integrity, err := suite.repo.VerifyIntegrity(suite.ctx, 1, 1000)
	require.NoError(suite.T(), err)

	for _, result := range results_integrity {
		assert.True(suite.T(), result.IsValid, "Audit log %d should be valid after concurrent creation", result.LogID)
	}
}

func TestAuditRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(AuditRepositoryTestSuite))
}
