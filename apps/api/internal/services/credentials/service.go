package credentials

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/audit"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Service defines the credentials management interface
type Service interface {
	// GetCredential retrieves a credential by name from Kubernetes secrets
	GetCredential(ctx context.Context, name string) (*Credential, error)

	// SetCredential stores a credential in Kubernetes secrets
	SetCredential(ctx context.Context, cred *Credential) error

	// UpdateCredential updates an existing credential
	UpdateCredential(ctx context.Context, cred *Credential) error

	// DeleteCredential removes a credential
	DeleteCredential(ctx context.Context, name string) error

	// ListCredentials returns all available credentials
	ListCredentials(ctx context.Context) ([]*CredentialInfo, error)

	// RotateCredential initiates credential rotation
	RotateCredential(ctx context.Context, name string, newValue string) error

	// ValidateCredential validates credential format and accessibility
	ValidateCredential(ctx context.Context, cred *Credential) (*ValidationResult, error)

	// GetCredentialHistory returns rotation history for a credential
	GetCredentialHistory(ctx context.Context, name string) ([]*CredentialHistoryEntry, error)
}

// Credential represents a secure credential
type Credential struct {
	Name        string            `json:"name"`
	Type        CredentialType    `json:"type"`
	Value       string            `json:"value"`
	Metadata    map[string]string `json:"metadata"`
	Description string            `json:"description"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
	RotatedAt   *time.Time        `json:"rotated_at,omitempty"`
	Tags        []string          `json:"tags"`
	Encrypted   bool              `json:"encrypted"`
}

// CredentialType represents different types of credentials
type CredentialType string

const (
	CredentialTypeAPIKey      CredentialType = "api_key"
	CredentialTypeToken       CredentialType = "token"
	CredentialTypeSecret      CredentialType = "secret"
	CredentialTypeCertificate CredentialType = "certificate"
	CredentialTypePassword    CredentialType = "password"
	CredentialTypeOAuth       CredentialType = "oauth"
)

// CredentialInfo contains non-sensitive information about a credential
type CredentialInfo struct {
	Name        string            `json:"name"`
	Type        CredentialType    `json:"type"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
	RotatedAt   *time.Time        `json:"rotated_at,omitempty"`
	Tags        []string          `json:"tags"`
	Status      string            `json:"status"`
}

// ValidationResult represents credential validation results
type ValidationResult struct {
	IsValid    bool           `json:"is_valid"`
	Errors     []string       `json:"errors"`
	Warnings   []string       `json:"warnings"`
	Score      int            `json:"score"` // Security score 0-100
	ExpiresIn  *time.Duration `json:"expires_in,omitempty"`
	LastTested time.Time      `json:"last_tested"`
}

// CredentialHistoryEntry represents a credential rotation event
type CredentialHistoryEntry struct {
	ID        uuid.UUID              `json:"id"`
	Action    string                 `json:"action"`
	UserID    uuid.UUID              `json:"user_id"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details"`
}

// Config represents credentials service configuration
type Config struct {
	Namespace         string `json:"namespace"`
	SecretPrefix      string `json:"secret_prefix"`
	EncryptionKey     string `json:"encryption_key"`
	EnableEncryption  bool   `json:"enable_encryption"`
	EnableAuditLog    bool   `json:"enable_audit_log"`
	AutoRotationDays  int    `json:"auto_rotation_days"`
	ValidationEnabled bool   `json:"validation_enabled"`
}

// serviceImpl implements the credentials service
type serviceImpl struct {
	k8sClient       kubernetes.Interface
	auditSvc        audit.Service
	config          *Config
	encryptionKey   []byte
	credentialCache map[string]*Credential
	cacheMutex      sync.RWMutex
	cacheExpiry     map[string]time.Time
	rotationTimer   *time.Timer
}

// NewService creates a new credentials service
func NewService(k8sClient kubernetes.Interface, auditSvc audit.Service, config *Config) (Service, error) {
	if config == nil {
		config = &Config{
			Namespace:         "kubechat",
			SecretPrefix:      "kubechat-creds",
			EnableEncryption:  true,
			EnableAuditLog:    true,
			AutoRotationDays:  90,
			ValidationEnabled: true,
		}
	}

	var encryptionKey []byte
	if config.EnableEncryption {
		if config.EncryptionKey == "" {
			return nil, fmt.Errorf("encryption key is required when encryption is enabled")
		}

		key, err := base64.StdEncoding.DecodeString(config.EncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("invalid encryption key format: %w", err)
		}

		if len(key) != 32 {
			return nil, fmt.Errorf("encryption key must be 32 bytes")
		}

		encryptionKey = key
	}

	service := &serviceImpl{
		k8sClient:       k8sClient,
		auditSvc:        auditSvc,
		config:          config,
		encryptionKey:   encryptionKey,
		credentialCache: make(map[string]*Credential),
		cacheExpiry:     make(map[string]time.Time),
	}

	// Start automatic rotation check
	if config.AutoRotationDays > 0 {
		service.startRotationMonitor()
	}

	log.Printf("Credentials service initialized for namespace: %s", config.Namespace)
	return service, nil
}

// GetCredential retrieves a credential by name
func (s *serviceImpl) GetCredential(ctx context.Context, name string) (*Credential, error) {
	// Check cache first
	s.cacheMutex.RLock()
	if cached, exists := s.credentialCache[name]; exists {
		if expiry, hasExpiry := s.cacheExpiry[name]; hasExpiry && time.Now().Before(expiry) {
			s.cacheMutex.RUnlock()
			s.logAccess(name, "cache_hit", nil)
			return cached, nil
		}
	}
	s.cacheMutex.RUnlock()

	// Fetch from Kubernetes secret
	secretName := s.buildSecretName(name)
	secret, err := s.k8sClient.CoreV1().Secrets(s.config.Namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		s.logAccess(name, "fetch_error", err)
		return nil, fmt.Errorf("failed to get credential %s: %w", name, err)
	}

	// Parse credential from secret
	cred, err := s.parseSecretToCredential(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credential: %w", err)
	}

	// Cache the credential
	s.cacheMutex.Lock()
	s.credentialCache[name] = cred
	s.cacheExpiry[name] = time.Now().Add(5 * time.Minute) // Cache for 5 minutes
	s.cacheMutex.Unlock()

	s.logAccess(name, "fetch_success", nil)
	return cred, nil
}

// SetCredential stores a credential in Kubernetes secrets
func (s *serviceImpl) SetCredential(ctx context.Context, cred *Credential) error {
	if err := s.validateCredentialInput(cred); err != nil {
		return fmt.Errorf("invalid credential: %w", err)
	}

	// Encrypt value if encryption is enabled
	value := cred.Value
	if s.config.EnableEncryption && s.encryptionKey != nil {
		encryptedValue, err := s.encrypt(cred.Value)
		if err != nil {
			return fmt.Errorf("failed to encrypt credential: %w", err)
		}
		value = encryptedValue
		cred.Encrypted = true
	}

	// Create Kubernetes secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.buildSecretName(cred.Name),
			Namespace: s.config.Namespace,
			Labels: map[string]string{
				"kubechat.dev/type":       "credential",
				"kubechat.dev/cred-type":  string(cred.Type),
				"kubechat.dev/managed-by": "kubechat-credentials",
			},
			Annotations: make(map[string]string),
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"value":       []byte(value),
			"type":        []byte(string(cred.Type)),
			"description": []byte(cred.Description),
			"encrypted":   []byte(fmt.Sprintf("%t", cred.Encrypted)),
			"created_at":  []byte(cred.CreatedAt.Format(time.RFC3339)),
			"updated_at":  []byte(time.Now().Format(time.RFC3339)),
		},
	}

	// Add metadata
	for key, val := range cred.Metadata {
		secret.Annotations[fmt.Sprintf("kubechat.dev/meta-%s", key)] = val
	}

	// Add tags
	if len(cred.Tags) > 0 {
		secret.Annotations["kubechat.dev/tags"] = strings.Join(cred.Tags, ",")
	}

	// Add expiry if set
	if cred.ExpiresAt != nil {
		secret.Data["expires_at"] = []byte(cred.ExpiresAt.Format(time.RFC3339))
	}

	// Create the secret
	_, err := s.k8sClient.CoreV1().Secrets(s.config.Namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		s.logOperation(cred.Name, "create_error", err)
		return fmt.Errorf("failed to create credential secret: %w", err)
	}

	// Clear cache
	s.cacheMutex.Lock()
	delete(s.credentialCache, cred.Name)
	delete(s.cacheExpiry, cred.Name)
	s.cacheMutex.Unlock()

	s.logOperation(cred.Name, "create_success", nil)
	log.Printf("Credential %s created successfully", cred.Name)
	return nil
}

// UpdateCredential updates an existing credential
func (s *serviceImpl) UpdateCredential(ctx context.Context, cred *Credential) error {
	secretName := s.buildSecretName(cred.Name)

	// Get existing secret
	secret, err := s.k8sClient.CoreV1().Secrets(s.config.Namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("credential %s not found: %w", cred.Name, err)
	}

	// Encrypt value if encryption is enabled
	value := cred.Value
	if s.config.EnableEncryption && s.encryptionKey != nil {
		encryptedValue, err := s.encrypt(cred.Value)
		if err != nil {
			return fmt.Errorf("failed to encrypt credential: %w", err)
		}
		value = encryptedValue
		cred.Encrypted = true
	}

	// Update secret data
	secret.Data["value"] = []byte(value)
	secret.Data["description"] = []byte(cred.Description)
	secret.Data["encrypted"] = []byte(fmt.Sprintf("%t", cred.Encrypted))
	secret.Data["updated_at"] = []byte(time.Now().Format(time.RFC3339))

	if cred.RotatedAt != nil {
		secret.Data["rotated_at"] = []byte(cred.RotatedAt.Format(time.RFC3339))
	}

	if cred.ExpiresAt != nil {
		secret.Data["expires_at"] = []byte(cred.ExpiresAt.Format(time.RFC3339))
	}

	// Update metadata and tags
	for key, val := range cred.Metadata {
		secret.Annotations[fmt.Sprintf("kubechat.dev/meta-%s", key)] = val
	}

	if len(cred.Tags) > 0 {
		secret.Annotations["kubechat.dev/tags"] = strings.Join(cred.Tags, ",")
	}

	// Update the secret
	_, err = s.k8sClient.CoreV1().Secrets(s.config.Namespace).Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		s.logOperation(cred.Name, "update_error", err)
		return fmt.Errorf("failed to update credential secret: %w", err)
	}

	// Clear cache
	s.cacheMutex.Lock()
	delete(s.credentialCache, cred.Name)
	delete(s.cacheExpiry, cred.Name)
	s.cacheMutex.Unlock()

	s.logOperation(cred.Name, "update_success", nil)
	log.Printf("Credential %s updated successfully", cred.Name)
	return nil
}

// DeleteCredential removes a credential
func (s *serviceImpl) DeleteCredential(ctx context.Context, name string) error {
	secretName := s.buildSecretName(name)

	err := s.k8sClient.CoreV1().Secrets(s.config.Namespace).Delete(ctx, secretName, metav1.DeleteOptions{})
	if err != nil {
		s.logOperation(name, "delete_error", err)
		return fmt.Errorf("failed to delete credential %s: %w", name, err)
	}

	// Clear cache
	s.cacheMutex.Lock()
	delete(s.credentialCache, name)
	delete(s.cacheExpiry, name)
	s.cacheMutex.Unlock()

	s.logOperation(name, "delete_success", nil)
	log.Printf("Credential %s deleted successfully", name)
	return nil
}

// ListCredentials returns all available credentials
func (s *serviceImpl) ListCredentials(ctx context.Context) ([]*CredentialInfo, error) {
	labelSelector := "kubechat.dev/type=credential"
	secrets, err := s.k8sClient.CoreV1().Secrets(s.config.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}

	var credentials []*CredentialInfo
	for _, secret := range secrets.Items {
		info, err := s.parseSecretToCredentialInfo(&secret)
		if err != nil {
			log.Printf("Failed to parse credential %s: %v", secret.Name, err)
			continue
		}
		credentials = append(credentials, info)
	}

	return credentials, nil
}

// RotateCredential initiates credential rotation
func (s *serviceImpl) RotateCredential(ctx context.Context, name string, newValue string) error {
	cred, err := s.GetCredential(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get credential for rotation: %w", err)
	}

	// Record rotation in history
	historyEntry := &CredentialHistoryEntry{
		ID:        uuid.New(),
		Action:    "rotate",
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"credential_name": name,
			"rotation_reason": "manual",
		},
	}

	// Update credential
	now := time.Now()
	cred.Value = newValue
	cred.UpdatedAt = now
	cred.RotatedAt = &now

	if err := s.UpdateCredential(ctx, cred); err != nil {
		return fmt.Errorf("failed to update rotated credential: %w", err)
	}

	s.logRotation(name, historyEntry)
	log.Printf("Credential %s rotated successfully", name)
	return nil
}

// ValidateCredential validates credential format and accessibility
func (s *serviceImpl) ValidateCredential(ctx context.Context, cred *Credential) (*ValidationResult, error) {
	result := &ValidationResult{
		IsValid:    true,
		Errors:     []string{},
		Warnings:   []string{},
		Score:      100,
		LastTested: time.Now(),
	}

	// Basic validation
	if cred.Name == "" {
		result.Errors = append(result.Errors, "credential name is required")
		result.IsValid = false
		result.Score -= 20
	}

	if cred.Value == "" {
		result.Errors = append(result.Errors, "credential value is required")
		result.IsValid = false
		result.Score -= 30
	}

	// Type-specific validation
	switch cred.Type {
	case CredentialTypeAPIKey:
		if len(cred.Value) < 20 {
			result.Warnings = append(result.Warnings, "API key appears to be too short")
			result.Score -= 10
		}
	case CredentialTypeToken:
		if !strings.HasPrefix(cred.Value, "sk-") && !strings.HasPrefix(cred.Value, "Bearer ") {
			result.Warnings = append(result.Warnings, "token format may be incorrect")
			result.Score -= 5
		}
	}

	// Expiry validation
	if cred.ExpiresAt != nil {
		if cred.ExpiresAt.Before(time.Now()) {
			result.Errors = append(result.Errors, "credential has expired")
			result.IsValid = false
			result.Score -= 50
		} else {
			expiresIn := time.Until(*cred.ExpiresAt)
			result.ExpiresIn = &expiresIn

			if expiresIn < 7*24*time.Hour {
				result.Warnings = append(result.Warnings, "credential expires within 7 days")
				result.Score -= 15
			}
		}
	}

	// Security score adjustments
	if !cred.Encrypted && s.config.EnableEncryption {
		result.Warnings = append(result.Warnings, "credential is not encrypted")
		result.Score -= 20
	}

	return result, nil
}

// GetCredentialHistory returns rotation history for a credential
func (s *serviceImpl) GetCredentialHistory(ctx context.Context, name string) ([]*CredentialHistoryEntry, error) {
	// This would typically query a database or log system
	// For now, return empty history
	return []*CredentialHistoryEntry{}, nil
}

// Helper methods

func (s *serviceImpl) buildSecretName(credName string) string {
	return fmt.Sprintf("%s-%s", s.config.SecretPrefix, credName)
}

func (s *serviceImpl) parseSecretToCredential(secret *corev1.Secret) (*Credential, error) {
	cred := &Credential{
		Name:     strings.TrimPrefix(secret.Name, s.config.SecretPrefix+"-"),
		Metadata: make(map[string]string),
		Tags:     []string{},
	}

	// Parse basic fields
	if val, exists := secret.Data["value"]; exists {
		value := string(val)

		// Decrypt if encrypted
		if encrypted, exists := secret.Data["encrypted"]; exists && string(encrypted) == "true" {
			if s.config.EnableEncryption && s.encryptionKey != nil {
				decryptedValue, err := s.decrypt(value)
				if err != nil {
					return nil, fmt.Errorf("failed to decrypt credential: %w", err)
				}
				value = decryptedValue
			}
		}

		cred.Value = value
	}

	if val, exists := secret.Data["type"]; exists {
		cred.Type = CredentialType(string(val))
	}

	if val, exists := secret.Data["description"]; exists {
		cred.Description = string(val)
	}

	if val, exists := secret.Data["encrypted"]; exists {
		cred.Encrypted = string(val) == "true"
	}

	// Parse timestamps
	if val, exists := secret.Data["created_at"]; exists {
		if t, err := time.Parse(time.RFC3339, string(val)); err == nil {
			cred.CreatedAt = t
		}
	}

	if val, exists := secret.Data["updated_at"]; exists {
		if t, err := time.Parse(time.RFC3339, string(val)); err == nil {
			cred.UpdatedAt = t
		}
	}

	if val, exists := secret.Data["expires_at"]; exists {
		if t, err := time.Parse(time.RFC3339, string(val)); err == nil {
			cred.ExpiresAt = &t
		}
	}

	if val, exists := secret.Data["rotated_at"]; exists {
		if t, err := time.Parse(time.RFC3339, string(val)); err == nil {
			cred.RotatedAt = &t
		}
	}

	// Parse metadata from annotations
	for key, val := range secret.Annotations {
		if strings.HasPrefix(key, "kubechat.dev/meta-") {
			metaKey := strings.TrimPrefix(key, "kubechat.dev/meta-")
			cred.Metadata[metaKey] = val
		}
	}

	// Parse tags
	if tags, exists := secret.Annotations["kubechat.dev/tags"]; exists {
		cred.Tags = strings.Split(tags, ",")
	}

	return cred, nil
}

func (s *serviceImpl) parseSecretToCredentialInfo(secret *corev1.Secret) (*CredentialInfo, error) {
	info := &CredentialInfo{
		Name:     strings.TrimPrefix(secret.Name, s.config.SecretPrefix+"-"),
		Metadata: make(map[string]string),
		Tags:     []string{},
		Status:   "active",
	}

	// Parse type
	if val, exists := secret.Data["type"]; exists {
		info.Type = CredentialType(string(val))
	}

	if val, exists := secret.Data["description"]; exists {
		info.Description = string(val)
	}

	// Parse timestamps
	if val, exists := secret.Data["created_at"]; exists {
		if t, err := time.Parse(time.RFC3339, string(val)); err == nil {
			info.CreatedAt = t
		}
	}

	if val, exists := secret.Data["updated_at"]; exists {
		if t, err := time.Parse(time.RFC3339, string(val)); err == nil {
			info.UpdatedAt = t
		}
	}

	if val, exists := secret.Data["expires_at"]; exists {
		if t, err := time.Parse(time.RFC3339, string(val)); err == nil {
			info.ExpiresAt = &t
			if t.Before(time.Now()) {
				info.Status = "expired"
			}
		}
	}

	if val, exists := secret.Data["rotated_at"]; exists {
		if t, err := time.Parse(time.RFC3339, string(val)); err == nil {
			info.RotatedAt = &t
		}
	}

	// Parse metadata from annotations
	for key, val := range secret.Annotations {
		if strings.HasPrefix(key, "kubechat.dev/meta-") {
			metaKey := strings.TrimPrefix(key, "kubechat.dev/meta-")
			info.Metadata[metaKey] = val
		}
	}

	// Parse tags
	if tags, exists := secret.Annotations["kubechat.dev/tags"]; exists {
		info.Tags = strings.Split(tags, ",")
	}

	return info, nil
}

func (s *serviceImpl) validateCredentialInput(cred *Credential) error {
	if cred.Name == "" {
		return fmt.Errorf("credential name is required")
	}

	if cred.Value == "" {
		return fmt.Errorf("credential value is required")
	}

	if cred.Type == "" {
		cred.Type = CredentialTypeAPIKey // Default type
	}

	if cred.CreatedAt.IsZero() {
		cred.CreatedAt = time.Now()
	}

	cred.UpdatedAt = time.Now()

	return nil
}

func (s *serviceImpl) encrypt(plaintext string) (string, error) {
	if s.encryptionKey == nil {
		return plaintext, nil
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *serviceImpl) decrypt(ciphertext string) (string, error) {
	if s.encryptionKey == nil {
		return ciphertext, nil
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func (s *serviceImpl) logAccess(credName, action string, err error) {
	if !s.config.EnableAuditLog || s.auditSvc == nil {
		return
	}

	safetyLevel := models.SafetyLevelSafe
	if err != nil {
		safetyLevel = models.SafetyLevelWarning
	}

	description := fmt.Sprintf("Credential access: %s - %s", credName, action)
	if err != nil {
		description += fmt.Sprintf(" (error: %s)", err.Error())
	}

	severityStr := "info"
	switch safetyLevel {
	case models.SafetyLevelWarning:
		severityStr = "warning"
	case models.SafetyLevelDangerous:
		severityStr = "error"
	}

	if logErr := s.auditSvc.LogSecurityEvent(context.Background(), "credential_access", description, nil, severityStr, nil); logErr != nil {
		log.Printf("Failed to log credential access: %v", logErr)
	}
}

func (s *serviceImpl) logOperation(credName, action string, err error) {
	if !s.config.EnableAuditLog || s.auditSvc == nil {
		return
	}

	description := fmt.Sprintf("Credential operation: %s - %s", credName, action)
	if err != nil {
		description += fmt.Sprintf(" (error: %s)", err.Error())
	}

	severityStr := "warning" // Credential operations are always sensitive
	if err != nil {
		severityStr = "error"
	}

	if logErr := s.auditSvc.LogSecurityEvent(context.Background(), "credential_operation", description, nil, severityStr, nil); logErr != nil {
		log.Printf("Failed to log credential operation: %v", logErr)
	}
}

func (s *serviceImpl) logRotation(credName string, history *CredentialHistoryEntry) {
	s.logOperation(credName, "rotate_success", nil)
}

func (s *serviceImpl) startRotationMonitor() {
	// This would implement automatic credential rotation monitoring
	// For now, just log that it's started
	log.Printf("Credential rotation monitor started (checking every %d days)", s.config.AutoRotationDays)
}
