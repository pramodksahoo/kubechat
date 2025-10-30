package poddisruptionbudgets

import (
	"encoding/json"
	"fmt"
	"net/http"

	policyV1 "k8s.io/api/policy/v1"

	"github.com/pramodksahoo/kubechat/backend/container"
	"github.com/pramodksahoo/kubechat/backend/handlers/base"
	"github.com/pramodksahoo/kubechat/backend/handlers/helpers"
	"github.com/labstack/echo/v4"
)

type PodDisruptionBudgetHandler struct {
	BaseHandler base.BaseHandler
}

func NewPodDisruptionBudgetRouteHandler(container container.Container, routeType base.RouteType) echo.HandlerFunc {
	return func(c echo.Context) error {
		handler := NewPodDisruptionBudgetHandler(c, container)

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

func NewPodDisruptionBudgetHandler(c echo.Context, container container.Container) *PodDisruptionBudgetHandler {
	config := c.QueryParam("config")
	cluster := c.QueryParam("cluster")

	informer := container.SharedInformerFactory(config, cluster).Policy().V1().PodDisruptionBudgets().Informer()
	informer.SetTransform(helpers.StripUnusedFields)

	handler := &PodDisruptionBudgetHandler{
		BaseHandler: base.BaseHandler{
			Kind:             "PodDisruptionBudget",
			Container:        container,
			Informer:         informer,
			RestClient:       container.ClientSet(config, cluster).PolicyV1().RESTClient(),
			QueryConfig:      config,
			QueryCluster:     cluster,
			InformerCacheKey: fmt.Sprintf("%s-%s-podDisruptionBudgetInformer", config, cluster),
			TransformFunc:    transformItems,
		},
	}
	cache := base.ResourceEventHandler[*policyV1.PodDisruptionBudget](&handler.BaseHandler)
	handler.BaseHandler.StartInformer(c, cache)
	handler.BaseHandler.WaitForSync(c)
	return handler
}

func transformItems(items []any, b *base.BaseHandler) ([]byte, error) {
	var list []policyV1.PodDisruptionBudget

	for _, obj := range items {
		if item, ok := obj.(*policyV1.PodDisruptionBudget); ok {
			list = append(list, *item)
		}
	}

	t := TransformPodDisruptionBudget(list)

	return json.Marshal(t)
}
