package manager

import (
	"net/http"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-infra/providers/httpserver"
)

func (p *provider) intRoutes(routes httpserver.Router) error {
	routes.GET("/api/logs-query/indices", p.inspectIndices)
	routes.POST("/api/logs-manager/indices/:addon", p.createByAddonIndex)
	return nil
}

func (p *provider) inspectIndices(r *http.Request) interface{} {
	return api.Success(p.indices.Load())
}

func (p *provider) createByAddonIndex(params struct {
	Addon string `param:"addon" validate:"required"`
}) interface{} {
	resp, err := p.createIndex(params.Addon)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(resp)
}
