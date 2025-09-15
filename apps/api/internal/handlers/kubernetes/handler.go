package kubernetes

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/pramodksahoo/kubechat/apps/api/internal/models"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/auth"
	"github.com/pramodksahoo/kubechat/apps/api/internal/services/kubernetes"
)

// Handler handles Kubernetes-related HTTP requests
type Handler struct {
	kubernetesService kubernetes.Service
}

// NewHandler creates a new Kubernetes handler
func NewHandler(kubernetesService kubernetes.Service) *Handler {
	return &Handler{
		kubernetesService: kubernetesService,
	}
}

// RegisterRoutes registers Kubernetes routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// All Kubernetes routes require authentication
	k8s := router.Group("/kubernetes")
	k8s.Use(auth.RequireWritePermission()) // Require at least user role for K8s operations
	{
		// Cluster information
		k8s.GET("/cluster", h.GetClusterInfo)
		k8s.GET("/health", h.HealthCheck)

		// Namespace operations
		k8s.GET("/namespaces", h.ListNamespaces)

		// Pod operations
		k8s.GET("/namespaces/:namespace/pods", h.ListPods)
		k8s.GET("/namespaces/:namespace/pods/:name", h.GetPod)
		k8s.DELETE("/namespaces/:namespace/pods/:name", h.DeletePod)
		k8s.GET("/namespaces/:namespace/pods/:name/logs", h.GetPodLogs)

		// Deployment operations
		k8s.GET("/namespaces/:namespace/deployments", h.ListDeployments)
		k8s.GET("/namespaces/:namespace/deployments/:name", h.GetDeployment)
		k8s.PUT("/namespaces/:namespace/deployments/:name/scale", h.ScaleDeployment)
		k8s.POST("/namespaces/:namespace/deployments/:name/restart", h.RestartDeployment)

		// Service operations
		k8s.GET("/namespaces/:namespace/services", h.ListServices)
		k8s.GET("/namespaces/:namespace/services/:name", h.GetService)

		// ConfigMap operations
		k8s.GET("/namespaces/:namespace/configmaps", h.ListConfigMaps)
		k8s.GET("/namespaces/:namespace/configmaps/:name", h.GetConfigMap)

		// Secret operations (metadata only)
		k8s.GET("/namespaces/:namespace/secrets", h.ListSecrets)

		// Operation validation
		k8s.POST("/validate", h.ValidateOperation)
	}
}

// GetClusterInfo returns cluster information
//
//	@Summary		Get Kubernetes cluster information
//	@Description	Returns basic information about the connected Kubernetes cluster
//	@Tags			Kubernetes Operations
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Cluster information"
//	@Failure		500	{object}	map[string]interface{}	"Failed to get cluster information"
//	@Security		BearerAuth
//	@Router			/kubernetes/cluster [get]
func (h *Handler) GetClusterInfo(c *gin.Context) {
	clusterInfo, err := h.kubernetesService.GetClusterInfo(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get cluster information",
			"code":  "CLUSTER_INFO_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": clusterInfo,
	})
}

// HealthCheck checks Kubernetes connectivity
//
//	@Summary		Check Kubernetes connectivity
//	@Description	Verifies connectivity to the Kubernetes cluster
//	@Tags			Kubernetes Operations
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Kubernetes cluster is accessible"
//	@Failure		503	{object}	map[string]interface{}	"Kubernetes cluster not accessible"
//	@Security		BearerAuth
//	@Router			/kubernetes/health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	err := h.kubernetesService.HealthCheck(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Kubernetes cluster not accessible",
			"code":  "K8S_HEALTH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "Kubernetes cluster is accessible",
	})
}

// ListNamespaces returns available namespaces
//
//	@Summary		List Kubernetes namespaces
//	@Description	Returns all available namespaces in the Kubernetes cluster
//	@Tags			Kubernetes Operations
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"List of namespaces"
//	@Failure		500	{object}	map[string]interface{}	"Failed to list namespaces"
//	@Security		BearerAuth
//	@Router			/kubernetes/namespaces [get]
func (h *Handler) ListNamespaces(c *gin.Context) {
	namespaces, err := h.kubernetesService.ListNamespaces(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list namespaces",
			"code":  "NAMESPACE_LIST_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": namespaces,
	})
}

// ListPods returns pods in a namespace
//
//	@Summary		List pods in namespace
//	@Description	Returns all pods in the specified namespace
//	@Tags			Kubernetes Operations
//	@Produce		json
//	@Param			namespace	path		string					true	"Namespace name"
//	@Success		200			{object}	map[string]interface{}	"List of pods"
//	@Failure		500			{object}	map[string]interface{}	"Failed to list pods"
//	@Security		BearerAuth
//	@Router			/kubernetes/namespaces/{namespace}/pods [get]
func (h *Handler) ListPods(c *gin.Context) {
	namespace := c.Param("namespace")

	pods, err := h.kubernetesService.ListPods(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list pods",
			"code":  "POD_LIST_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": pods,
	})
}

// GetPod returns a specific pod
//
//	@Summary		Get specific pod
//	@Description	Returns detailed information about a specific pod
//	@Tags			Kubernetes Operations
//	@Produce		json
//	@Param			namespace	path		string					true	"Namespace name"
//	@Param			name		path		string					true	"Pod name"
//	@Success		200			{object}	map[string]interface{}	"Pod information"
//	@Failure		404			{object}	map[string]interface{}	"Pod not found"
//	@Security		BearerAuth
//	@Router			/kubernetes/namespaces/{namespace}/pods/{name} [get]
func (h *Handler) GetPod(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	pod, err := h.kubernetesService.GetPod(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Pod not found",
			"code":  "POD_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": pod,
	})
}

// DeletePod deletes a pod
//
//	@Summary		Delete a pod
//	@Description	Deletes a specific pod from the cluster (dangerous operation)
//	@Tags			Kubernetes Operations
//	@Produce		json
//	@Param			namespace	path		string					true	"Namespace name"
//	@Param			name		path		string					true	"Pod name"
//	@Success		200			{object}	map[string]interface{}	"Pod deletion initiated"
//	@Failure		403			{object}	map[string]interface{}	"Operation not allowed"
//	@Failure		500			{object}	map[string]interface{}	"Failed to delete pod"
//	@Security		BearerAuth
//	@Router			/kubernetes/namespaces/{namespace}/pods/{name} [delete]
func (h *Handler) DeletePod(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	// Create and log the operation for audit purposes
	userID, _, _, _ := auth.ExtractUserContext(c)
	operation := &models.KubernetesOperation{
		ID:        uuid.New(),
		UserID:    uuid.MustParse(userID),
		Operation: "delete",
		Resource:  "pods",
		Namespace: namespace,
		Name:      name,
	}

	// Validate the operation
	if err := h.kubernetesService.ValidateOperation(c.Request.Context(), operation); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
			"code":  "OPERATION_NOT_ALLOWED",
		})
		return
	}

	err := h.kubernetesService.DeletePod(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete pod",
			"code":  "POD_DELETE_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pod deletion initiated",
		"pod": gin.H{
			"name":      name,
			"namespace": namespace,
		},
	})
}

// GetPodLogs returns pod logs
//
//	@Summary		Get pod logs
//	@Description	Returns logs from a specific pod and container
//	@Tags			Kubernetes Operations
//	@Produce		json
//	@Param			namespace	path		string					true	"Namespace name"
//	@Param			name		path		string					true	"Pod name"
//	@Param			tail_lines	query		int						false	"Number of lines to tail"	default(100)
//	@Param			timestamps	query		bool					false	"Include timestamps in logs"
//	@Param			container	query		string					false	"Specific container name"
//	@Success		200			{object}	map[string]interface{}	"Pod logs"
//	@Failure		500			{object}	map[string]interface{}	"Failed to get pod logs"
//	@Security		BearerAuth
//	@Router			/kubernetes/namespaces/{namespace}/pods/{name}/logs [get]
func (h *Handler) GetPodLogs(c *gin.Context) {
	namespace := c.Param("namespace")
	podName := c.Param("name")

	// Parse query parameters
	options := &models.LogOptions{
		TailLines:  100, // Default
		Timestamps: false,
	}

	if tailLinesStr := c.Query("tail_lines"); tailLinesStr != "" {
		if tailLines, err := strconv.ParseInt(tailLinesStr, 10, 64); err == nil {
			options.TailLines = tailLines
		}
	}

	if c.Query("timestamps") == "true" {
		options.Timestamps = true
	}

	if container := c.Query("container"); container != "" {
		options.Container = container
	}

	logs, err := h.kubernetesService.GetPodLogs(c.Request.Context(), namespace, podName, options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get pod logs",
			"code":  "POD_LOGS_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"pod":       podName,
			"namespace": namespace,
			"logs":      logs,
		},
	})
}

// ListDeployments returns deployments in a namespace
//
//	@Summary		List deployments in namespace
//	@Description	Returns all deployments in the specified namespace
//	@Tags			Kubernetes Operations
//	@Produce		json
//	@Param			namespace	path		string					true	"Namespace name"
//	@Success		200			{object}	map[string]interface{}	"List of deployments"
//	@Failure		500			{object}	map[string]interface{}	"Failed to list deployments"
//	@Security		BearerAuth
//	@Router			/kubernetes/namespaces/{namespace}/deployments [get]
func (h *Handler) ListDeployments(c *gin.Context) {
	namespace := c.Param("namespace")

	deployments, err := h.kubernetesService.ListDeployments(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list deployments",
			"code":  "DEPLOYMENT_LIST_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": deployments,
	})
}

// GetDeployment returns a specific deployment
//
//	@Summary		Get specific deployment
//	@Description	Returns detailed information about a specific deployment
//	@Tags			Kubernetes Operations
//	@Produce		json
//	@Param			namespace	path		string					true	"Namespace name"
//	@Param			name		path		string					true	"Deployment name"
//	@Success		200			{object}	map[string]interface{}	"Deployment information"
//	@Failure		404			{object}	map[string]interface{}	"Deployment not found"
//	@Security		BearerAuth
//	@Router			/kubernetes/namespaces/{namespace}/deployments/{name} [get]
func (h *Handler) GetDeployment(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	deployment, err := h.kubernetesService.GetDeployment(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Deployment not found",
			"code":  "DEPLOYMENT_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": deployment,
	})
}

// ScaleDeployment scales a deployment
//
//	@Summary		Scale a deployment
//	@Description	Scales a deployment to the specified number of replicas
//	@Tags			Kubernetes Operations
//	@Accept			json
//	@Produce		json
//	@Param			namespace	path		string					true	"Namespace name"
//	@Param			name		path		string					true	"Deployment name"
//	@Param			request		body		object{replicas=int}	true	"Scaling request"	example({"replicas": 3})
//	@Success		200			{object}	map[string]interface{}	"Deployment scaling initiated"
//	@Failure		400			{object}	map[string]interface{}	"Invalid request format"
//	@Failure		403			{object}	map[string]interface{}	"Operation not allowed"
//	@Failure		500			{object}	map[string]interface{}	"Failed to scale deployment"
//	@Security		BearerAuth
//	@Router			/kubernetes/namespaces/{namespace}/deployments/{name}/scale [put]
func (h *Handler) ScaleDeployment(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	var req struct {
		Replicas int32 `json:"replicas" binding:"required,min=0,max=10"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Create and validate the operation
	userID, _, _, _ := auth.ExtractUserContext(c)
	operation := &models.KubernetesOperation{
		ID:        uuid.New(),
		UserID:    uuid.MustParse(userID),
		Operation: "scale",
		Resource:  "deployments",
		Namespace: namespace,
		Name:      name,
	}

	if err := h.kubernetesService.ValidateOperation(c.Request.Context(), operation); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
			"code":  "OPERATION_NOT_ALLOWED",
		})
		return
	}

	err := h.kubernetesService.ScaleDeployment(c.Request.Context(), namespace, name, req.Replicas)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to scale deployment",
			"code":  "DEPLOYMENT_SCALE_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Deployment scaling initiated",
		"deployment": gin.H{
			"name":      name,
			"namespace": namespace,
			"replicas":  req.Replicas,
		},
	})
}

// RestartDeployment restarts a deployment
//
//	@Summary		Restart a deployment
//	@Description	Restarts all pods in a deployment by triggering a rollout
//	@Tags			Kubernetes Operations
//	@Produce		json
//	@Param			namespace	path		string					true	"Namespace name"
//	@Param			name		path		string					true	"Deployment name"
//	@Success		200			{object}	map[string]interface{}	"Deployment restart initiated"
//	@Failure		403			{object}	map[string]interface{}	"Operation not allowed"
//	@Failure		500			{object}	map[string]interface{}	"Failed to restart deployment"
//	@Security		BearerAuth
//	@Router			/kubernetes/namespaces/{namespace}/deployments/{name}/restart [post]
func (h *Handler) RestartDeployment(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	// Create and validate the operation
	userID, _, _, _ := auth.ExtractUserContext(c)
	operation := &models.KubernetesOperation{
		ID:        uuid.New(),
		UserID:    uuid.MustParse(userID),
		Operation: "restart",
		Resource:  "deployments",
		Namespace: namespace,
		Name:      name,
	}

	if err := h.kubernetesService.ValidateOperation(c.Request.Context(), operation); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
			"code":  "OPERATION_NOT_ALLOWED",
		})
		return
	}

	err := h.kubernetesService.RestartDeployment(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to restart deployment",
			"code":  "DEPLOYMENT_RESTART_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Deployment restart initiated",
		"deployment": gin.H{
			"name":      name,
			"namespace": namespace,
		},
	})
}

// ListServices returns services in a namespace
//
//	@Summary		List services in namespace
//	@Description	Returns all services in the specified namespace
//	@Tags			Kubernetes Operations
//	@Produce		json
//	@Param			namespace	path		string					true	"Namespace name"
//	@Success		200			{object}	map[string]interface{}	"List of services"
//	@Failure		500			{object}	map[string]interface{}	"Failed to list services"
//	@Security		BearerAuth
//	@Router			/kubernetes/namespaces/{namespace}/services [get]
func (h *Handler) ListServices(c *gin.Context) {
	namespace := c.Param("namespace")

	services, err := h.kubernetesService.ListServices(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list services",
			"code":  "SERVICE_LIST_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": services,
	})
}

// GetService returns a specific service
//
//	@Summary		Get specific service
//	@Description	Returns detailed information about a specific service
//	@Tags			Kubernetes Operations
//	@Produce		json
//	@Param			namespace	path		string					true	"Namespace name"
//	@Param			name		path		string					true	"Service name"
//	@Success		200			{object}	map[string]interface{}	"Service information"
//	@Failure		404			{object}	map[string]interface{}	"Service not found"
//	@Security		BearerAuth
//	@Router			/kubernetes/namespaces/{namespace}/services/{name} [get]
func (h *Handler) GetService(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	service, err := h.kubernetesService.GetService(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Service not found",
			"code":  "SERVICE_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": service,
	})
}

// ListConfigMaps returns configmaps in a namespace
//
//	@Summary		List ConfigMaps in namespace
//	@Description	Returns all ConfigMaps in the specified namespace
//	@Tags			Kubernetes Operations
//	@Produce		json
//	@Param			namespace	path		string					true	"Namespace name"
//	@Success		200			{object}	map[string]interface{}	"List of ConfigMaps"
//	@Failure		500			{object}	map[string]interface{}	"Failed to list ConfigMaps"
//	@Security		BearerAuth
//	@Router			/kubernetes/namespaces/{namespace}/configmaps [get]
func (h *Handler) ListConfigMaps(c *gin.Context) {
	namespace := c.Param("namespace")

	configMaps, err := h.kubernetesService.ListConfigMaps(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list configmaps",
			"code":  "CONFIGMAP_LIST_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": configMaps,
	})
}

// GetConfigMap returns a specific configmap
//
//	@Summary		Get specific ConfigMap
//	@Description	Returns detailed information about a specific ConfigMap
//	@Tags			Kubernetes Operations
//	@Produce		json
//	@Param			namespace	path		string					true	"Namespace name"
//	@Param			name		path		string					true	"ConfigMap name"
//	@Success		200			{object}	map[string]interface{}	"ConfigMap information"
//	@Failure		404			{object}	map[string]interface{}	"ConfigMap not found"
//	@Security		BearerAuth
//	@Router			/kubernetes/namespaces/{namespace}/configmaps/{name} [get]
func (h *Handler) GetConfigMap(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	configMap, err := h.kubernetesService.GetConfigMap(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "ConfigMap not found",
			"code":  "CONFIGMAP_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": configMap,
	})
}

// ListSecrets returns secrets in a namespace (metadata only)
//
//	@Summary		List Secrets in namespace
//	@Description	Returns metadata for all Secrets in the namespace (data is not exposed)
//	@Tags			Kubernetes Operations
//	@Produce		json
//	@Param			namespace	path		string					true	"Namespace name"
//	@Success		200			{object}	map[string]interface{}	"List of Secret metadata"
//	@Failure		500			{object}	map[string]interface{}	"Failed to list Secrets"
//	@Security		BearerAuth
//	@Router			/kubernetes/namespaces/{namespace}/secrets [get]
func (h *Handler) ListSecrets(c *gin.Context) {
	namespace := c.Param("namespace")

	secrets, err := h.kubernetesService.ListSecrets(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list secrets",
			"code":  "SECRET_LIST_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": secrets,
	})
}

// ValidateOperation validates a Kubernetes operation
//
//	@Summary		Validate Kubernetes operation
//	@Description	Validates if a Kubernetes operation is allowed and safe to execute
//	@Tags			Kubernetes Operations
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.KubernetesOperation	true	"Operation to validate"
//	@Success		200		{object}	map[string]interface{}		"Operation validation result"
//	@Failure		400		{object}	map[string]interface{}		"Invalid request format or validation failed"
//	@Failure		403		{object}	map[string]interface{}		"Operation not allowed"
//	@Security		BearerAuth
//	@Router			/kubernetes/validate [post]
func (h *Handler) ValidateOperation(c *gin.Context) {
	var operation models.KubernetesOperation
	if err := c.ShouldBindJSON(&operation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Set user context
	userID, _, _, _ := auth.ExtractUserContext(c)
	operation.UserID = uuid.MustParse(userID)
	operation.ID = uuid.New()

	// Validate the operation
	if err := operation.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_FAILED",
		})
		return
	}

	// Check service-level validation
	if err := h.kubernetesService.ValidateOperation(c.Request.Context(), &operation); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
			"code":  "OPERATION_NOT_ALLOWED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"operation": gin.H{
			"id":           operation.ID,
			"description":  operation.GetDescription(),
			"safety_level": operation.GetSafetyLevel(),
		},
	})
}
