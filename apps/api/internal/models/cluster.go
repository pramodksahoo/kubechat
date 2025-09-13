package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ClusterConfig represents a Kubernetes cluster configuration with encryption
type ClusterConfig struct {
	ID            uuid.UUID              `json:"id" db:"id"`
	UserID        uuid.UUID              `json:"user_id" db:"user_id"`
	ClusterName   string                 `json:"cluster_name" db:"cluster_name"`
	ClusterConfig map[string]interface{} `json:"cluster_config" db:"cluster_config"`
	IsActive      bool                   `json:"is_active" db:"is_active"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

// ClusterConfigRequest represents the data needed to create/update a cluster config
type ClusterConfigRequest struct {
	ClusterName   string                 `json:"cluster_name" validate:"required,min=1,max=255"`
	ClusterConfig map[string]interface{} `json:"cluster_config" validate:"required"`
	IsActive      *bool                  `json:"is_active,omitempty"`
}

// EncryptedClusterConfig represents the encrypted version stored in database
type EncryptedClusterConfig struct {
	ID              uuid.UUID `json:"id" db:"id"`
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	ClusterName     string    `json:"cluster_name" db:"cluster_name"`
	EncryptedConfig string    `json:"-" db:"cluster_config"` // Encrypted JSON as string
	IsActive        bool      `json:"is_active" db:"is_active"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// ClusterEncryption handles AES-256 encryption for cluster configurations
type ClusterEncryption struct {
	key []byte
}

// NewClusterEncryption creates a new encryption instance with the provided key
func NewClusterEncryption(key []byte) (*ClusterEncryption, error) {
	if len(key) != 32 { // AES-256 requires 32-byte key
		return nil, errors.New("encryption key must be 32 bytes for AES-256")
	}
	return &ClusterEncryption{key: key}, nil
}

// Encrypt encrypts the cluster configuration using AES-256-GCM
func (ce *ClusterEncryption) Encrypt(config map[string]interface{}) (string, error) {
	// Convert config to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	// Create AES cipher
	block, err := aes.NewCipher(ce.key)
	if err != nil {
		return "", err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, configJSON, nil)

	// Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts the cluster configuration using AES-256-GCM
func (ce *ClusterEncryption) Decrypt(encryptedConfig string) (map[string]interface{}, error) {
	// Decode from base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedConfig)
	if err != nil {
		return nil, err
	}

	// Create AES cipher
	block, err := aes.NewCipher(ce.key)
	if err != nil {
		return nil, err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var config map[string]interface{}
	if err := json.Unmarshal(plaintext, &config); err != nil {
		return nil, err
	}

	return config, nil
}

// NewClusterConfig creates a new cluster configuration
func NewClusterConfig(userID uuid.UUID, req ClusterConfigRequest) *ClusterConfig {
	isActive := false
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	return &ClusterConfig{
		ID:            uuid.New(),
		UserID:        userID,
		ClusterName:   req.ClusterName,
		ClusterConfig: req.ClusterConfig,
		IsActive:      isActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// EncryptConfig encrypts the cluster configuration for storage
func (cc *ClusterConfig) EncryptConfig(encryption *ClusterEncryption) (*EncryptedClusterConfig, error) {
	encryptedConfig, err := encryption.Encrypt(cc.ClusterConfig)
	if err != nil {
		return nil, err
	}

	return &EncryptedClusterConfig{
		ID:              cc.ID,
		UserID:          cc.UserID,
		ClusterName:     cc.ClusterName,
		EncryptedConfig: encryptedConfig,
		IsActive:        cc.IsActive,
		CreatedAt:       cc.CreatedAt,
		UpdatedAt:       cc.UpdatedAt,
	}, nil
}

// DecryptConfig decrypts the cluster configuration from storage
func (ecc *EncryptedClusterConfig) DecryptConfig(encryption *ClusterEncryption) (*ClusterConfig, error) {
	config, err := encryption.Decrypt(ecc.EncryptedConfig)
	if err != nil {
		return nil, err
	}

	return &ClusterConfig{
		ID:            ecc.ID,
		UserID:        ecc.UserID,
		ClusterName:   ecc.ClusterName,
		ClusterConfig: config,
		IsActive:      ecc.IsActive,
		CreatedAt:     ecc.CreatedAt,
		UpdatedAt:     ecc.UpdatedAt,
	}, nil
}

// Update updates the cluster configuration
func (cc *ClusterConfig) Update(req ClusterConfigRequest) {
	cc.ClusterName = req.ClusterName
	cc.ClusterConfig = req.ClusterConfig
	if req.IsActive != nil {
		cc.IsActive = *req.IsActive
	}
	cc.UpdatedAt = time.Now()
}

// IsValidKubeConfig performs basic validation on kubeconfig structure
func (cc *ClusterConfig) IsValidKubeConfig() bool {
	// Check for required kubeconfig fields
	if cc.ClusterConfig == nil {
		return false
	}

	// Check for basic kubeconfig structure
	clusters, hasClusters := cc.ClusterConfig["clusters"]
	contexts, hasContexts := cc.ClusterConfig["contexts"]
	users, hasUsers := cc.ClusterConfig["users"]

	return hasClusters && hasContexts && hasUsers &&
		clusters != nil && contexts != nil && users != nil
}

// GetAPIServerURL extracts the API server URL from kubeconfig
func (cc *ClusterConfig) GetAPIServerURL() string {
	if !cc.IsValidKubeConfig() {
		return ""
	}

	clusters, ok := cc.ClusterConfig["clusters"].([]interface{})
	if !ok || len(clusters) == 0 {
		return ""
	}

	firstCluster, ok := clusters[0].(map[string]interface{})
	if !ok {
		return ""
	}

	cluster, ok := firstCluster["cluster"].(map[string]interface{})
	if !ok {
		return ""
	}

	server, ok := cluster["server"].(string)
	if !ok {
		return ""
	}

	return server
}

// SanitizedConfig returns a sanitized version of the config for logging (without sensitive data)
func (cc *ClusterConfig) SanitizedConfig() map[string]interface{} {
	sanitized := make(map[string]interface{})

	// Only include non-sensitive metadata
	sanitized["cluster_name"] = cc.ClusterName
	sanitized["api_server"] = cc.GetAPIServerURL()
	sanitized["is_active"] = cc.IsActive
	sanitized["created_at"] = cc.CreatedAt
	sanitized["updated_at"] = cc.UpdatedAt

	return sanitized
}
