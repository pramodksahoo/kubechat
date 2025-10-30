package ingresses

import (
	"encoding/json"
	"fmt"
	"net/http"

	networkingV1 "k8s.io/api/networking/v1"

	"github.com/pramodksahoo/kubechat/backend/container"
	"github.com/pramodksahoo/kubechat/backend/handlers/base"
	"github.com/pramodksahoo/kubechat/backend/handlers/helpers"
	"github.com/labstack/echo/v4"
)

type IngressHandler struct {
	BaseHandler base.BaseHandler
}

func NewIngressRouteHandler(container container.Container, routeType base.RouteType) echo.HandlerFunc {
	return func(c echo.Context) error {
		handler := NewIngressHandler(c, container)

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

func NewIngressHandler(c echo.Context, container container.Container) *IngressHandler {
	config := c.QueryParam("config")
	cluster := c.QueryParam("cluster")

	informer := container.SharedInformerFactory(config, cluster).Networking().V1().Ingresses().Informer()
	informer.SetTransform(helpers.StripUnusedFields)

	handler := &IngressHandler{
		BaseHandler: base.BaseHandler{
			Kind:             "Ingress",
			Container:        container,
			Informer:         informer,
			RestClient:       container.ClientSet(config, cluster).NetworkingV1().RESTClient(),
			QueryConfig:      config,
			QueryCluster:     cluster,
			InformerCacheKey: fmt.Sprintf("%s-%s-IngressInformer", config, cluster),
			TransformFunc:    transformItems,
		},
	}
	cache := base.ResourceEventHandler[*networkingV1.Ingress](&handler.BaseHandler)
	handler.BaseHandler.StartInformer(c, cache)
	handler.BaseHandler.WaitForSync(c)
	return handler
}

func transformItems(items []any, b *base.BaseHandler) ([]byte, error) {
	var list []networkingV1.Ingress

	for _, obj := range items {
		if item, ok := obj.(*networkingV1.Ingress); ok {
			list = append(list, *item)
		}
	}
	t := TransformIngress(list)

	return json.Marshal(t)
}
