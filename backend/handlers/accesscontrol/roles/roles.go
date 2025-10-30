package roles

import (
	"encoding/json"
	"fmt"
	"net/http"

	rbacV1 "k8s.io/api/rbac/v1"

	"github.com/pramodksahoo/kubechat/backend/container"
	"github.com/pramodksahoo/kubechat/backend/handlers/base"
	"github.com/pramodksahoo/kubechat/backend/handlers/helpers"
	"github.com/labstack/echo/v4"
)

type RolesHandler struct {
	BaseHandler base.BaseHandler
}

func NewRoleRouteHandler(container container.Container, routeType base.RouteType) echo.HandlerFunc {
	return func(c echo.Context) error {
		handler := NewRolesHandler(c, container)

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

func NewRolesHandler(c echo.Context, container container.Container) *RolesHandler {
	config := c.QueryParam("config")
	cluster := c.QueryParam("cluster")

	informer := container.SharedInformerFactory(config, cluster).Rbac().V1().Roles().Informer()
	informer.SetTransform(helpers.StripUnusedFields)

	handler := &RolesHandler{
		BaseHandler: base.BaseHandler{
			Kind:             "Role",
			Container:        container,
			Informer:         informer,
			RestClient:       container.ClientSet(config, cluster).RbacV1().RESTClient(),
			QueryConfig:      config,
			QueryCluster:     cluster,
			InformerCacheKey: fmt.Sprintf("%s-%s-roleInformer", config, cluster),
			TransformFunc:    transformItems,
		},
	}
	cache := base.ResourceEventHandler[*rbacV1.Role](&handler.BaseHandler)
	handler.BaseHandler.StartInformer(c, cache)
	handler.BaseHandler.WaitForSync(c)
	return handler
}

func transformItems(items []any, b *base.BaseHandler) ([]byte, error) {
	list := make([]rbacV1.Role, 0)

	for _, obj := range items {
		if item, ok := obj.(*rbacV1.Role); ok {
			list = append(list, *item)
		}
	}
	t := TransformRoleList(list)

	return json.Marshal(t)
}
