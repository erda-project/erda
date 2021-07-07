// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package collector

import (
	"context"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/erda-project/erda-infra/providers/httpserver"
)

func (p *provider) intRouteV2(r httpserver.Router) error {
	// standard API version two
	signAuth := p.authSignedRequest()
	limit := p.BodySizeLimit()
	r.POST("/api/v2/collect/logs/:source", p.collectLogsV2, limit, signAuth)
	return nil
}

func (p *provider) collectLogsV2(ctx echo.Context) error {
	source := ctx.Param("source")
	if source == "" {
		return ctx.NoContent(http.StatusBadRequest)
	}
	name := source + "_log"

	if v := ctx.Request().Header.Get("Content-Type"); v != "application/x-protobuf" {
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported Content-Type:"+v)
	}

	body, err := ReadRequestBody(ctx.Request())
	if err != nil {
		p.Logger.Errorf("fail to read request body, err: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "fail to read request body")
	}

	vctx := context.WithValue(context.TODO(), "topic", name)
	if err := p.output.Send(vctx, body); err != nil {
		p.Logger.Errorf("fail to send msg to kafka, name: %s, err: %v", name, err)
	}
	return ctx.NoContent(http.StatusNoContent)

}

func (p *provider) BodySizeLimit() interface{} {
	return middleware.BodyLimitWithConfig(middleware.BodyLimitConfig{
		Limit: p.Cfg.Limiter.RequestBodySize,
	})
}
