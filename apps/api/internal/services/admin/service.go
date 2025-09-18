package admin

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/repositories"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type AdminService interface {
	EnsureAdminUser(ctx context.Context) error
	SyncPasswordFromSecret(ctx context.Context) error
}

type service struct {
	userRepo   repositories.UserRepository
	k8sClient  kubernetes.Interface
	namespace  string
	secretName string
}

type Config struct {
	Namespace  string
	SecretName string
}

func NewService(userRepo repositories.UserRepository, config *Config) (AdminService, error) {
	if config.Namespace == "" {
		config.Namespace = "kubechat"
	}
	if config.SecretName == "" {
		config.SecretName = "kubechat-admin-secret"
	}

	k8sClient, err := createK8sClient()
	if err != nil {
		log.Printf("Warning: Failed to create Kubernetes client: %v", err)
		log.Println("Admin service will read credentials from mounted volume")
		k8sClient = nil
	}

	return &service{
		userRepo:   userRepo,
		k8sClient:  k8sClient,
		namespace:  config.Namespace,
		secretName: config.SecretName,
	}, nil
}

func createK8sClient() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return clientset, nil
}

func (s *service) EnsureAdminUser(ctx context.Context) error {
	log.Println("Checking for admin user existence...")

	existingAdmin, err := s.userRepo.GetByUsername(ctx, "admin")
	if err != nil {
		return fmt.Errorf("failed to check for existing admin user: %w", err)
	}

	if existingAdmin != nil {
		log.Println("Admin user already exists, syncing password from secret...")
		return s.SyncPasswordFromSecret(ctx)
	}

	log.Println("Admin user does not exist, creating new admin user...")

	secretData, err := s.getAdminSecret(ctx)
	if err != nil {
		return fmt.Errorf("failed to get admin secret: %w", err)
	}

	adminUser := &models.User{
		ID:        uuid.New(),
		Username:  "admin",
		Email:     secretData.Email,
		Role:      models.RoleAdmin,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := adminUser.HashPassword(secretData.Password); err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	if err := s.userRepo.Create(ctx, adminUser); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	log.Println("✅ Admin user created successfully with credentials from K8s Secret")
	return nil
}

func (s *service) SyncPasswordFromSecret(ctx context.Context) error {
	log.Println("Syncing admin password from K8s Secret...")

	secretData, err := s.getAdminSecret(ctx)
	if err != nil {
		return fmt.Errorf("failed to get admin secret: %w", err)
	}

	adminUser, err := s.userRepo.GetByUsername(ctx, "admin")
	if err != nil {
		return fmt.Errorf("failed to get admin user: %w", err)
	}

	if adminUser == nil {
		return fmt.Errorf("admin user not found")
	}

	if err := adminUser.CheckPassword(secretData.Password); err == nil {
		log.Println("Admin password is already in sync with secret")
		return nil
	}

	log.Println("Admin password differs from secret, updating database...")

	if err := adminUser.HashPassword(secretData.Password); err != nil {
		return fmt.Errorf("failed to hash new admin password: %w", err)
	}

	if err := s.userRepo.Update(ctx, adminUser); err != nil {
		return fmt.Errorf("failed to update admin user password: %w", err)
	}

	log.Println("✅ Admin password synced successfully from K8s Secret")
	return nil
}

type AdminSecretData struct {
	Username string
	Password string
	Email    string
}

func (s *service) getAdminSecret(ctx context.Context) (*AdminSecretData, error) {
	// Try to read from mounted volume first (preferred method)
	secretData, err := s.getAdminSecretFromVolume()
	if err == nil {
		return secretData, nil
	}
	log.Printf("Warning: Failed to read admin secret from volume: %v", err)

	// Fallback to Kubernetes API if available
	if s.k8sClient != nil {
		secretData, err := s.getAdminSecretFromAPI(ctx)
		if err == nil {
			return secretData, nil
		}
		log.Printf("Warning: Failed to read admin secret from API: %v", err)
	}

	return nil, fmt.Errorf("failed to read admin secret from both volume and API")
}

func (s *service) getAdminSecretFromVolume() (*AdminSecretData, error) {
	const secretPath = "/etc/kubechat/admin"

	readFile := func(filename string) (string, error) {
		data, err := os.ReadFile(filepath.Join(secretPath, filename))
		if err != nil {
			return "", fmt.Errorf("failed to read %s: %w", filename, err)
		}
		return string(data), nil
	}

	username, err := readFile("admin_username")
	if err != nil {
		return nil, err
	}

	password, err := readFile("admin_password")
	if err != nil {
		return nil, err
	}

	email, err := readFile("admin_email")
	if err != nil {
		return nil, err
	}

	// When secrets are mounted as volumes, Kubernetes automatically decodes base64
	// So we don't need to decode again - the data is already in plain text
	return &AdminSecretData{
		Username: username,
		Password: password,
		Email:    email,
	}, nil
}

func (s *service) getAdminSecretFromAPI(ctx context.Context) (*AdminSecretData, error) {
	secret, err := s.k8sClient.CoreV1().Secrets(s.namespace).Get(ctx, s.secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret %s/%s: %w", s.namespace, s.secretName, err)
	}

	decode := func(key string) (string, error) {
		data, exists := secret.Data[key]
		if !exists {
			return "", fmt.Errorf("key %s not found in secret", key)
		}
		return string(data), nil
	}

	username, err := decode("admin_username")
	if err != nil {
		return nil, err
	}

	password, err := decode("admin_password")
	if err != nil {
		return nil, err
	}

	email, err := decode("admin_email")
	if err != nil {
		return nil, err
	}

	decodedUsername, err := base64.StdEncoding.DecodeString(username)
	if err != nil {
		return nil, fmt.Errorf("failed to decode username: %w", err)
	}

	decodedPassword, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		return nil, fmt.Errorf("failed to decode password: %w", err)
	}

	decodedEmail, err := base64.StdEncoding.DecodeString(email)
	if err != nil {
		return nil, fmt.Errorf("failed to decode email: %w", err)
	}

	return &AdminSecretData{
		Username: string(decodedUsername),
		Password: string(decodedPassword),
		Email:    string(decodedEmail),
	}, nil
}
