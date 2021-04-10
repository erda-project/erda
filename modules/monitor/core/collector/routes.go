// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package collector

import (
	"bufio"
	"fmt"
	"hash/adler32"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/rakyll/statik/fs"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/providers/httpserver"
	_ "github.com/erda-project/erda/modules/monitor/core/collector/statik" // include static files
)

//go:generate statik -src=./ -ns "monitor/metrics-collector" -include=*.jpg,*.txt,*.html,*.css,*.js
func (c *collector) intRoutes(routes httpserver.Router) error {
	assets, err := fs.NewWithNamespace("monitor/metrics-collector")
	if err != nil {
		return fmt.Errorf("fail to init file system: %s", err)
	}
	fileSystem := http.FileSystem(assets)
	routes.File("/ta.js", "/ta.js", httpserver.WithFileSystem(fileSystem))

	// browser and mobile metrics
	routes.POST("/collect", c.collectAnalytics) // compatible for legacy
	routes.POST("/collect/analytics", c.collectAnalytics)

	// logs and metrics
	basicAuth := c.basicAuth()
	routes.POST("/collect/logs/:source", c.collectLogs, basicAuth)
	routes.POST("/collect/:metric", c.collectMetric, basicAuth)

	routes.POST("/collect/notify-metrics", c.collectNotifyMetric, basicAuth)
	return nil
}

func (c *collector) basicAuth() interface{} {
	return middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
		Validator: func(username string, password string, context echo.Context) (bool, error) {
			if username == c.Cfg.Auth.Username && password == c.Cfg.Auth.Password {
				return true, nil
			}
			return false, nil
		},
		Skipper: func(context echo.Context) bool {
			if c.Cfg.Auth.Force {
				return false
			}
			// 兼容旧版本，没有添加认证的客户端，这个版本先跳过
			authorizationHeader := context.Request().Header.Get("Authorization")
			return authorizationHeader == ""
		},
	})
}

func (c *collector) collectLogs(ctx echo.Context) error {
	source := ctx.Param("source")
	if source == "" {
		return ctx.NoContent(http.StatusBadRequest)
	}
	name := source + "_log"

	body, err := ReadRequestBody(ctx.Request())
	if err != nil {
		logrus.Errorf("fail to read request body, err: %v", err)
		return err
	}
	if isJSONArray(body) {
		logrus.Warningf("the body is not a json array. body=%s", string(body))
		return ctx.NoContent(http.StatusNoContent)
	}

	if _, err := jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			logrus.Errorf("fail to json parse, err: %v", err)
			return
		}
		if err := c.send(name, value); err != nil {
			logrus.Errorf("fail to send msg to kafka, name: %s, err: %v", name, err)
		}
	}); err != nil {
		return err
	}
	return ctx.NoContent(http.StatusNoContent)
}

func (c *collector) collectAnalytics(ctx echo.Context) error {
	params, err := ctx.FormParams()
	if err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}
	ak := ctx.FormValue("ak") // ak as terminus key
	ai := ctx.FormValue("ai")

	if c.Cfg.TaSamplingRate <= 0 {
		return ctx.NoContent(http.StatusOK)
	}
	if c.Cfg.TaSamplingRate < 99.999999 { // 99.999999认为是100，避免浮点数精度问题导致 100.0<100.0 为true
		hash := float64(adler32.Checksum([]byte(ak)) % 100)
		if c.Cfg.TaSamplingRate <= hash {
			return ctx.NoContent(http.StatusOK)
		}
	}
	cid := ctx.FormValue("cid")
	if ak == "" || cid == "" {
		return ctx.String(http.StatusBadRequest, "ak and cid are required")
	}
	uid := ctx.FormValue("uid")

	var ip string
	// k8s 的一个bug，使用 X-Original-Forwarded-For 而不是 X-Forwarded-For
	ipAddrs := ctx.Request().Header.Get("X-Original-Forwarded-For")
	if ipAddrs != "" {
		ip = strings.TrimSpace(strings.SplitN(ipAddrs, ",", 2)[0])
	} else {
		ip = ctx.RealIP()
	}

	// Data from Client
	for _, v := range params["data"] {
		qs, err := url.ParseQuery(v)
		if err != nil {
			return ctx.NoContent(http.StatusBadRequest)
		}
		qs.Set("date", strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10))

		if qs.Get("ua") == "" {
			ua := ctx.Request().Header.Get("User-Agent")
			qs.Set("ua", ua)
		}
		var message string
		if len(ai) > 0 {
			message = fmt.Sprintf(`%s,%s,%s,%s,%s,%s`,
				ak, cid, uid, ip, qs.Encode(), ai)
		} else {
			message = fmt.Sprintf(`%s,%s,%s,%s,%s`,
				ak, cid, uid, ip, qs.Encode())
		}
		_ = c.send("analytics", []byte(message))
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (c *collector) collectMetric(ctx echo.Context) error {
	contentType := ctx.Request().Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		return c.parseJSON(ctx, ctx.Param("metric"))
	}
	return c.parseLine(ctx, ctx.Param("metric"))
}

func (c *collector) collectNotifyMetric(ctx echo.Context) error {
	contentType := ctx.Request().Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		return c.parseJSON(ctx, ctx.Param("erda-notify-event"))
	}
	return nil
}

func (c *collector) parseJSON(ctx echo.Context, name string) error {
	body, err := ReadRequestBody(ctx.Request())
	if err != nil {
		return err
	}
	if _, err := jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err == nil {
			err = c.send(name, value)
		}
	}, name); err != nil {
		return err
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (c *collector) parseLine(ctx echo.Context, name string) error {
	reader, err := ReadRequestBodyReader(ctx.Request())
	if err != nil {
		return err
	}
	buf := bufio.NewReader(reader)
	for {
		line, err := buf.ReadString('\n')
		if e := c.send(name, []byte(line)); e != nil {
			return e
		}
		if err != nil || err == io.EOF {
			break
		}
	}
	return ctx.NoContent(http.StatusNoContent)
}
