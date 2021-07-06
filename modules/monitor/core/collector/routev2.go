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
	"net/http"

	"github.com/buger/jsonparser"
	"github.com/labstack/echo"

	"github.com/erda-project/erda-infra/providers/httpserver"
	v2 "github.com/erda-project/erda/modules/monitor/core/collector/v2"
)

func (c *collector) intRouteV2(r httpserver.Router) error {
	// standard API version two
	signAuth := c.authSignedRequest()
	r.POST("/api/v2/collect/logs/:source", c.collectLogsV2, signAuth)

	return nil
}

func (c *collector) collectLogsV2(ctx echo.Context) error {
	source := ctx.Param("source")
	if source == "" {
		return ctx.NoContent(http.StatusBadRequest)
	}
	name := source + "_log"

	body, err := v2.ReadRequestBody(ctx.Request())
	if err != nil {
		c.Logger.Errorf("fail to read request body, err: %v", err)
		return err
	}

	// data type handler
	switch ctx.Request().Header.Get("Content-Type") {
	case "application/json":
		if !isJSONArray(body) {
			c.Logger.Warnf("the body is not a json array. body=%s", string(body))
			return ctx.NoContent(http.StatusNoContent)
		}

		if _, err := jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			if err != nil {
				c.Logger.Errorf("fail to json parse, err: %v", err)
				return
			}
			if err := c.send(name, value); err != nil {
				c.Logger.Errorf("fail to send msg to kafka, name: %s, err: %v", name, err)
			}
		}); err != nil {
			return err
		}
		return ctx.NoContent(http.StatusNoContent)
	case "application/x-protobuf":
		// TODO
		return ctx.NoContent(http.StatusNoContent)
	default:
		c.Logger.Errorf("unsupported Content-Type: %s", ctx.Request().Header.Get("Content-Type"))
		return err
	}
}
