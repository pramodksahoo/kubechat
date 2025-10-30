package serviceaccounts

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pramodksahoo/kubechat/backend/container"
	"github.com/pramodksahoo/kubechat/backend/handlers/base"
	"github.com/pramodksahoo/kubechat/backend/handlers/helpers"
	"github.com/labstack/echo/v4"
	coreV1 "k8s.io/api/core/v1"
)

type ServiceAccountsHandler struct {
	BaseHandler base.BaseHandler
}

func NewServiceAccountsRouteHandler(container container.Container, routeType base.RouteType) echo.HandlerFunc {
	return func(c echo.Context) error {
		handler := NewServiceAccountsHandler(c, container)

		switch routeType {
		case base.GetList:
			return handler.BaseHandler.GetList(c)
		case base.GetDetails:
			return handler.BaseHandler.GetDetails(c)
		case base.GetEvents:
			return handler.BaseHandler.GetEvents(c)
		case base.GetYaml:
			return handler.BaseHandler.GetYaml(c)
		case base.Delete:
			return handler.BaseHandler.Delete(c)
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, "Unknown route type")
		}
	}
}

func NewServiceAccountsHandler(c echo.Context, container container.Container) *ServiceAccountsHandler {
	config := c.QueryParam("config")
	cluster := c.QueryParam("cluster")

	informer := container.SharedInformerFactory(config, cluster).Core().V1().ServiceAccounts().Informer()
	informer.SetTransform(helpers.StripUnusedFields)

	handler := &ServiceAccountsHandler{
		BaseHandler: base.BaseHandler{
			Kind:             "ServiceAccount",
			Container:        container,
			Informer:         informer,
			RestClient:       container.ClientSet(config, cluster).CoreV1().RESTClient(),
			QueryConfig:      config,
			QueryCluster:     cluster,
			InformerCacheKey: fmt.Sprintf("%s-%s-serviceAccountInformer", config, cluster),
			TransformFunc:    transformItems,
		},
	}
	cache := base.ResourceEventHandler[*coreV1.ServiceAccount](&handler.BaseHandler)
	handler.BaseHandler.StartInformer(c, cache)
	handler.BaseHandler.WaitForSync(c)
	return handler
}

func transformItems(items []any, b *base.BaseHandler) ([]byte, error) {
	var list []coreV1.ServiceAccount

	for _, obj := range items {
		if item, ok := obj.(*coreV1.ServiceAccount); ok {
			list = append(list, *item)
		}
	}
	t := TransformServiceAccountsList(list)

	return json.Marshal(t)
}
