package replicaset

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pramodksahoo/kubechat/backend/handlers/base"
	"github.com/pramodksahoo/kubechat/backend/handlers/helpers"

	"github.com/pramodksahoo/kubechat/backend/container"
	"github.com/labstack/echo/v4"
	appV1 "k8s.io/api/apps/v1"
)

type ReplicaSetHandler struct {
	BaseHandler base.BaseHandler
}

func NewReplicaSetRouteHandler(container container.Container, routeType base.RouteType) echo.HandlerFunc {
	return func(c echo.Context) error {
		handler := NewReplicaSetHandler(c, container)

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

func NewReplicaSetHandler(c echo.Context, container container.Container) *ReplicaSetHandler {
	config := c.QueryParam("config")
	cluster := c.QueryParam("cluster")

	informer := container.SharedInformerFactory(config, cluster).Apps().V1().ReplicaSets().Informer()
	informer.SetTransform(helpers.StripUnusedFields)

	handler := &ReplicaSetHandler{
		BaseHandler: base.BaseHandler{
			Kind:             "Replicaset",
			Container:        container,
			Informer:         informer,
			RestClient:       container.ClientSet(config, cluster).AppsV1().RESTClient(),
			QueryConfig:      config,
			QueryCluster:     cluster,
			InformerCacheKey: fmt.Sprintf("%s-%s-replicaSetInformer", config, cluster),
			TransformFunc:    transformItems,
		},
	}

	cache := base.ResourceEventHandler[*appV1.ReplicaSet](&handler.BaseHandler)
	handler.BaseHandler.StartInformer(c, cache)
	handler.BaseHandler.WaitForSync(c)
	return handler
}

func transformItems(items []any, b *base.BaseHandler) ([]byte, error) {
	var replicasSets []appV1.ReplicaSet

	for _, obj := range items {
		if rep, ok := obj.(*appV1.ReplicaSet); ok {
			replicasSets = append(replicasSets, *rep)
		}
	}

	t := TransformReplicaSetList(replicasSets)

	return json.Marshal(t)
}
