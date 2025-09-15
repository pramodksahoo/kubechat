package models_test

import (
	"crypto/rand"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
)

func TestNewClusterEncryption(t *testing.T) {
	tests := []struct {
		name    string
		keySize int
		wantErr bool
	}{
		{
			name:    "valid 32-byte key",
			keySize: 32,
			wantErr: false,
		},
		{
			name:    "invalid 16-byte key",
			keySize: 16,
			wantErr: true,
		},
		{
			name:    "invalid 64-byte key",
			keySize: 64,
			wantErr: true,
		},
		{
			name:    "invalid empty key",
			keySize: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := make([]byte, tt.keySize)
			if tt.keySize > 0 {
				_, err := rand.Read(key)
				require.NoError(t, err)
			}

			encryption, err := models.NewClusterEncryption(key)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, encryption)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, encryption)
			}
		})
	}
}

func TestClusterEncryption_EncryptDecrypt(t *testing.T) {
	// Generate a valid 32-byte key
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	encryption, err := models.NewClusterEncryption(key)
	require.NoError(t, err)

	tests := []struct {
		name   string
		config map[string]interface{}
	}{
		{
			name: "simple config",
			config: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Config",
				"clusters": []map[string]interface{}{
					{
						"name": "test-cluster",
						"cluster": map[string]interface{}{
							"server": "https://kubernetes.example.com",
						},
					},
				},
			},
		},
		{
			name: "complex kubeconfig",
			config: map[string]interface{}{
				"apiVersion":      "v1",
				"kind":            "Config",
				"current-context": "test-context",
				"clusters": []map[string]interface{}{
					{
						"name": "test-cluster",
						"cluster": map[string]interface{}{
							"server":                   "https://kubernetes.example.com:6443",
							"certificate-authority":    "LS0tLS1CRUdJTi...",
							"insecure-skip-tls-verify": false,
						},
					},
				},
				"contexts": []map[string]interface{}{
					{
						"name": "test-context",
						"context": map[string]interface{}{
							"cluster":   "test-cluster",
							"user":      "test-user",
							"namespace": "default",
						},
					},
				},
				"users": []map[string]interface{}{
					{
						"name": "test-user",
						"user": map[string]interface{}{
							"token": "eyJhbGciOiJSUzI1NiIs...",
						},
					},
				},
			},
		},
		{
			name:   "empty config",
			config: map[string]interface{}{},
		},
		{
			name: "config with special characters",
			config: map[string]interface{}{
				"special": "!@#$%^&*()_+-=[]{}|;':\",./<>?",
				"unicode": "测试数据🔐",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt the config
			encryptedData, err := encryption.Encrypt(tt.config)
			require.NoError(t, err)
			assert.NotEmpty(t, encryptedData)

			// Encrypted data should be base64 encoded
			assert.Regexp(t, "^[A-Za-z0-9+/=]+$", encryptedData)

			// Decrypt the config
			decryptedConfig, err := encryption.Decrypt(encryptedData)
			require.NoError(t, err)

			// Compare original and decrypted configs
			// Note: JSON marshaling/unmarshaling can change slice types, so we compare the JSON representations
			originalJSON, _ := json.Marshal(tt.config)
			decryptedJSON, _ := json.Marshal(decryptedConfig)
			assert.JSONEq(t, string(originalJSON), string(decryptedJSON))
		})
	}
}

func TestClusterEncryption_DecryptInvalidData(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	encryption, err := models.NewClusterEncryption(key)
	require.NoError(t, err)

	tests := []struct {
		name          string
		encryptedData string
		wantErr       bool
	}{
		{
			name:          "invalid base64",
			encryptedData: "invalid-base64-data!",
			wantErr:       true,
		},
		{
			name:          "empty string",
			encryptedData: "",
			wantErr:       true,
		},
		{
			name:          "too short ciphertext",
			encryptedData: "YWJjZA==", // "abcd" in base64, too short for GCM
			wantErr:       true,
		},
		{
			name:          "corrupted data",
			encryptedData: "dGVzdGRhdGF0aGF0aXN0b29zaG9ydA==", // valid base64 but invalid ciphertext
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := encryption.Decrypt(tt.encryptedData)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClusterEncryption_DifferentKeys(t *testing.T) {
	// Create two different encryption instances with different keys
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	_, err := rand.Read(key1)
	require.NoError(t, err)
	_, err = rand.Read(key2)
	require.NoError(t, err)

	encryption1, err := models.NewClusterEncryption(key1)
	require.NoError(t, err)
	encryption2, err := models.NewClusterEncryption(key2)
	require.NoError(t, err)

	config := map[string]interface{}{
		"test": "data",
	}

	// Encrypt with first key
	encryptedData, err := encryption1.Encrypt(config)
	require.NoError(t, err)

	// Try to decrypt with second key - should fail
	_, err = encryption2.Decrypt(encryptedData)
	assert.Error(t, err)
}

func TestNewClusterConfig(t *testing.T) {
	userID := uuid.New()
	req := models.ClusterConfigRequest{
		ClusterName: "test-cluster",
		ClusterConfig: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Config",
		},
		IsActive: nil, // Test default
	}

	config := models.NewClusterConfig(userID, req)

	assert.NotEqual(t, uuid.Nil, config.ID)
	assert.Equal(t, userID, config.UserID)
	assert.Equal(t, req.ClusterName, config.ClusterName)
	assert.Equal(t, req.ClusterConfig, config.ClusterConfig)
	assert.False(t, config.IsActive) // Default should be false
	assert.False(t, config.CreatedAt.IsZero())
	assert.False(t, config.UpdatedAt.IsZero())

	// Test with IsActive set to true
	activeTrue := true
	req.IsActive = &activeTrue
	config2 := models.NewClusterConfig(userID, req)
	assert.True(t, config2.IsActive)
}

func TestClusterConfig_EncryptDecryptConfig(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	encryption, err := models.NewClusterEncryption(key)
	require.NoError(t, err)

	config := &models.ClusterConfig{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		ClusterName: "test-cluster",
		ClusterConfig: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Config",
			"clusters": []map[string]interface{}{
				{
					"name": "test-cluster",
					"cluster": map[string]interface{}{
						"server": "https://kubernetes.example.com",
					},
				},
			},
		},
		IsActive: true,
	}

	// Test encryption
	encryptedConfig, err := config.EncryptConfig(encryption)
	require.NoError(t, err)
	assert.NotNil(t, encryptedConfig)
	assert.Equal(t, config.ID, encryptedConfig.ID)
	assert.Equal(t, config.UserID, encryptedConfig.UserID)
	assert.Equal(t, config.ClusterName, encryptedConfig.ClusterName)
	assert.Equal(t, config.IsActive, encryptedConfig.IsActive)
	assert.NotEmpty(t, encryptedConfig.EncryptedConfig)

	// Test decryption
	decryptedConfig, err := encryptedConfig.DecryptConfig(encryption)
	require.NoError(t, err)
	assert.Equal(t, config.ID, decryptedConfig.ID)
	assert.Equal(t, config.UserID, decryptedConfig.UserID)
	assert.Equal(t, config.ClusterName, decryptedConfig.ClusterName)
	// Use JSON comparison for ClusterConfig since JSON marshal/unmarshal can change slice types
	originalJSON, _ := json.Marshal(config.ClusterConfig)
	decryptedJSON, _ := json.Marshal(decryptedConfig.ClusterConfig)
	assert.JSONEq(t, string(originalJSON), string(decryptedJSON))
	assert.Equal(t, config.IsActive, decryptedConfig.IsActive)
}

func TestClusterConfig_Update(t *testing.T) {
	config := &models.ClusterConfig{
		ID:            uuid.New(),
		UserID:        uuid.New(),
		ClusterName:   "old-cluster",
		ClusterConfig: map[string]interface{}{"old": "config"},
		IsActive:      false,
	}

	oldUpdatedAt := config.UpdatedAt

	// Test update with new values
	activeTrue := true
	req := models.ClusterConfigRequest{
		ClusterName:   "new-cluster",
		ClusterConfig: map[string]interface{}{"new": "config"},
		IsActive:      &activeTrue,
	}

	config.Update(req)

	assert.Equal(t, req.ClusterName, config.ClusterName)
	assert.Equal(t, req.ClusterConfig, config.ClusterConfig)
	assert.True(t, config.IsActive)
	assert.True(t, config.UpdatedAt.After(oldUpdatedAt))

	// Test update with nil IsActive (should not change)
	req2 := models.ClusterConfigRequest{
		ClusterName:   "newer-cluster",
		ClusterConfig: map[string]interface{}{"newer": "config"},
		IsActive:      nil,
	}

	config.Update(req2)
	assert.True(t, config.IsActive) // Should remain true
}

func TestClusterConfig_IsValidKubeConfig(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]interface{}
		valid  bool
	}{
		{
			name: "valid kubeconfig",
			config: map[string]interface{}{
				"clusters": []interface{}{},
				"contexts": []interface{}{},
				"users":    []interface{}{},
			},
			valid: true,
		},
		{
			name: "missing clusters",
			config: map[string]interface{}{
				"contexts": []interface{}{},
				"users":    []interface{}{},
			},
			valid: false,
		},
		{
			name: "missing contexts",
			config: map[string]interface{}{
				"clusters": []interface{}{},
				"users":    []interface{}{},
			},
			valid: false,
		},
		{
			name: "missing users",
			config: map[string]interface{}{
				"clusters": []interface{}{},
				"contexts": []interface{}{},
			},
			valid: false,
		},
		{
			name:   "nil config",
			config: nil,
			valid:  false,
		},
		{
			name:   "empty config",
			config: map[string]interface{}{},
			valid:  false,
		},
		{
			name: "null values",
			config: map[string]interface{}{
				"clusters": nil,
				"contexts": []interface{}{},
				"users":    []interface{}{},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &models.ClusterConfig{
				ClusterConfig: tt.config,
			}
			assert.Equal(t, tt.valid, config.IsValidKubeConfig())
		})
	}
}

func TestClusterConfig_GetAPIServerURL(t *testing.T) {
	tests := []struct {
		name     string
		config   map[string]interface{}
		expected string
	}{
		{
			name: "valid cluster with server",
			config: map[string]interface{}{
				"clusters": []interface{}{
					map[string]interface{}{
						"cluster": map[string]interface{}{
							"server": "https://kubernetes.example.com:6443",
						},
					},
				},
				"contexts": []interface{}{},
				"users":    []interface{}{},
			},
			expected: "https://kubernetes.example.com:6443",
		},
		{
			name: "invalid config",
			config: map[string]interface{}{
				"invalid": "config",
			},
			expected: "",
		},
		{
			name: "empty clusters",
			config: map[string]interface{}{
				"clusters": []interface{}{},
				"contexts": []interface{}{},
				"users":    []interface{}{},
			},
			expected: "",
		},
		{
			name: "malformed cluster",
			config: map[string]interface{}{
				"clusters": []interface{}{
					"invalid-cluster-format",
				},
				"contexts": []interface{}{},
				"users":    []interface{}{},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &models.ClusterConfig{
				ClusterConfig: tt.config,
			}
			result := config.GetAPIServerURL()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClusterConfig_SanitizedConfig(t *testing.T) {
	config := &models.ClusterConfig{
		ClusterName: "test-cluster",
		ClusterConfig: map[string]interface{}{
			"clusters": []interface{}{
				map[string]interface{}{
					"cluster": map[string]interface{}{
						"server": "https://kubernetes.example.com:6443",
					},
				},
			},
			"contexts": []interface{}{},
			"users":    []interface{}{},
		},
		IsActive: true,
	}

	sanitized := config.SanitizedConfig()

	// Should contain non-sensitive metadata
	assert.Equal(t, config.ClusterName, sanitized["cluster_name"])
	assert.Equal(t, config.GetAPIServerURL(), sanitized["api_server"])
	assert.Equal(t, config.IsActive, sanitized["is_active"])
	assert.Contains(t, sanitized, "created_at")
	assert.Contains(t, sanitized, "updated_at")

	// Should not contain sensitive cluster config data
	assert.NotContains(t, sanitized, "clusters")
	assert.NotContains(t, sanitized, "users")
	assert.NotContains(t, sanitized, "contexts")
}

func BenchmarkClusterEncryption_Encrypt(b *testing.B) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(b, err)

	encryption, err := models.NewClusterEncryption(key)
	require.NoError(b, err)

	config := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Config",
		"clusters":   []interface{}{},
		"contexts":   []interface{}{},
		"users":      []interface{}{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encryption.Encrypt(config)
	}
}

func BenchmarkClusterEncryption_Decrypt(b *testing.B) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(b, err)

	encryption, err := models.NewClusterEncryption(key)
	require.NoError(b, err)

	config := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Config",
		"clusters":   []interface{}{},
		"contexts":   []interface{}{},
		"users":      []interface{}{},
	}

	// Pre-encrypt for benchmark
	encryptedData, err := encryption.Encrypt(config)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encryption.Decrypt(encryptedData)
	}
}
