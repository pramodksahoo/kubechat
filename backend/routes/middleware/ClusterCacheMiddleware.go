package middleware

import (
	"fmt"

	"github.com/pramodksahoo/kubechat/backend/container"
	"github.com/pramodksahoo/kubechat/backend/handlers/config/secrets"
	"github.com/pramodksahoo/kubechat/backend/handlers/helpers"
	"github.com/pramodksahoo/kubechat/backend/handlers/workloads/deployments"
	"github.com/labstack/echo/v4"

	"github.com/pramodksahoo/kubechat/backend/handlers/accesscontrol/clusterroles"
	clusterrolebindings "github.com/pramodksahoo/kubechat/backend/handlers/accesscontrol/clusterrolesbindings"
	"github.com/pramodksahoo/kubechat/backend/handlers/accesscontrol/roles"
	rolebindings "github.com/pramodksahoo/kubechat/backend/handlers/accesscontrol/rolesbindings"
	"github.com/pramodksahoo/kubechat/backend/handlers/accesscontrol/serviceaccounts"
	configmaps "github.com/pramodksahoo/kubechat/backend/handlers/config/configMaps"
	horizontalpodautoscalers "github.com/pramodksahoo/kubechat/backend/handlers/config/horizontalPodAutoscalers"
	"github.com/pramodksahoo/kubechat/backend/handlers/config/leases"
	limitranges "github.com/pramodksahoo/kubechat/backend/handlers/config/limitRanges"
	poddisruptionbudgets "github.com/pramodksahoo/kubechat/backend/handlers/config/podDisruptionBudgets"
	priorityclasses "github.com/pramodksahoo/kubechat/backend/handlers/config/priorityClasses"
	resourcequotas "github.com/pramodksahoo/kubechat/backend/handlers/config/resourceQuotas"
	runtimeclasses "github.com/pramodksahoo/kubechat/backend/handlers/config/runtimeClasses"
	"github.com/pramodksahoo/kubechat/backend/handlers/events"
	"github.com/pramodksahoo/kubechat/backend/handlers/namespaces"
	"github.com/pramodksahoo/kubechat/backend/handlers/network/endpoints"
	"github.com/pramodksahoo/kubechat/backend/handlers/network/ingresses"
	"github.com/pramodksahoo/kubechat/backend/handlers/network/services"
	"github.com/pramodksahoo/kubechat/backend/handlers/nodes"
	"github.com/pramodksahoo/kubechat/backend/handlers/storage/persistentvolumeclaims"
	"github.com/pramodksahoo/kubechat/backend/handlers/storage/persistentvolumes"
	"github.com/pramodksahoo/kubechat/backend/handlers/storage/storageclasses"
	cronjobs "github.com/pramodksahoo/kubechat/backend/handlers/workloads/cronJobs"
	"github.com/pramodksahoo/kubechat/backend/handlers/workloads/daemonsets"
	"github.com/pramodksahoo/kubechat/backend/handlers/workloads/jobs"
	"github.com/pramodksahoo/kubechat/backend/handlers/workloads/pods"
	"github.com/pramodksahoo/kubechat/backend/handlers/workloads/replicaset"
	statefulset "github.com/pramodksahoo/kubechat/backend/handlers/workloads/statefulsets"
)

func ClusterCacheMiddleware(container container.Container) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if shouldSkip(c) {
				return next(c)
			}

			config := c.QueryParam("config")
			cluster := c.QueryParam("cluster")

			allResourcesKey := fmt.Sprintf(helpers.AllResourcesCacheKeyFormat, config, cluster)

			conn := container.Config().KubeConfig[config].Clusters[cluster]
			if !conn.IsConnected() {
				conn.MarkAsConnected()
			}

			_, exists := container.Cache().GetIfPresent(allResourcesKey)
			if !exists {
				helpers.CacheAllResources(container, config, cluster)
				loadAllInformerOfCluster(c, container)
			}

			return next(c)
		}
	}
}

func loadAllInformerOfCluster(c echo.Context, container container.Container) {
	go pods.NewPodsHandler(c, container)
	go deployments.NewDeploymentsHandler(c, container)
	go daemonsets.NewDaemonSetsHandler(c, container)
	go replicaset.NewReplicaSetHandler(c, container)
	go statefulset.NewSatefulSetHandler(c, container)
	go cronjobs.NewCronJobsHandler(c, container)
	go jobs.NewJobsHandler(c, container)

	// Storage
	go persistentvolumeclaims.NewPersistentVolumeClaimsHandler(c, container)
	go persistentvolumes.NewPersistentVolumeHandler(c, container)
	go storageclasses.NewStorageClassesHandler(c, container)

	// Config
	go configmaps.NewConfigMapsHandler(c, container)
	go secrets.NewSecretsHandler(c, container)
	go resourcequotas.NewResourceQuotaHandler(c, container)
	go namespaces.NewNamespacesHandler(c, container)
	go horizontalpodautoscalers.NewHorizontalPodAutoScalerHandler(c, container)
	go poddisruptionbudgets.NewPodDisruptionBudgetHandler(c, container)
	go priorityclasses.NewPriorityClassHandler(c, container)
	go runtimeclasses.NewRunTimeClassHandler(c, container)
	go leases.NewLeasesHandler(c, container)

	// AccessControl
	go serviceaccounts.NewServiceAccountsHandler(c, container)
	go roles.NewRolesHandler(c, container)
	go rolebindings.NewRoleBindingHandler(c, container)
	go clusterroles.NewRolesHandler(c, container)
	go clusterrolebindings.NewClusterRoleBindingHandler(c, container)

	// Network
	go endpoints.NewEndpointsHandler(c, container)
	go ingresses.NewIngressHandler(c, container)
	go services.NewServicesHandler(c, container)
	go limitranges.NewLimitRangesHandler(c, container)

	go nodes.NewNodeHandler(c, container)
	go events.NewEventsHandler(c, container)
}
