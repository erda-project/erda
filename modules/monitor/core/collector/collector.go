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
	"fmt"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	writer "github.com/erda-project/erda-infra/pkg/parallel-writer"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/kafka"
)

type config struct {
	Auth struct {
		Username string `file:"username"`
		Password string `file:"password"`
		Force    bool   `file:"force"`
	}
	Output         kafka.ProducerConfig `file:"output"`
	TaSamplingRate float64              `file:"ta_sampling_rate" default:"100"`
}

type define struct{}

func (m *define) Services() []string     { return []string{"metrics-collector"} }
func (m *define) Dependencies() []string { return []string{"http-server", "kafka-producer"} }
func (m *define) Summary() string        { return "log and metrics collector" }
func (m *define) Description() string    { return m.Summary() }
func (m *define) Config() interface{}    { return &config{} }
func (m *define) Creator() servicehub.Creator {
	return func() servicehub.Provider { return &collector{} }
}

// collector .
type collector struct {
	Cfg    *config
	Logger logs.Logger
	writer writer.Writer
	kafka  kafka.Interface
}

func (c *collector) Init(ctx servicehub.Context) error {
	c.kafka = ctx.Service("kafka-producer").(kafka.Interface)
	w, err := c.kafka.NewProducer(&c.Cfg.Output)
	if err != nil {
		return fmt.Errorf("fail to create kafka producer: %s", err)
	}
	c.writer = w

	routes := ctx.Service("http-server",
		// telemetry.HttpMetric(),
		c.logrusMiddleware(),
		interceptors.CORS(),
		interceptors.Recover(c.Logger),
	).(httpserver.Router)
	err = c.intRoutes(routes)
	if err != nil {
		return fmt.Errorf("fail to init routes: %s", err)
	}
	return nil
}

// `${time_rfc3339} ${remote_ip} ${header:terminus-request-id} ${method} ${host} ${uri} ${status} ${latency_human} in ${bytes_in} out ${bytes_out} ${error}` + "\n"
func (c *collector) logrusMiddleware(skipper ...middleware.Skipper) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) (err error) {
			req := ctx.Request()
			res := ctx.Response()
			start := time.Now()
			if err = next(ctx); err != nil {
				ctx.Error(err)
			}
			stop := time.Now()
			statusCode := res.Status
			if stop.Sub(start).Milliseconds() > 50 || statusCode >= 300 {
				errorMessage := ""
				if err != nil {
					errorMessage = err.Error()
				}
				c.Logger.Warnf("%s %s %s %s %s %d %s %s %s",
					ctx.RealIP(),
					req.Header.Get("terminus-request-id"),
					req.Method,
					req.Host,
					req.RequestURI,
					statusCode,
					stop.Sub(start).String(),
					req.Header.Get("Custom-Content-Encoding"),
					errorMessage)
			}
			return err
		}
	}
}

func init() {
	servicehub.RegisterProvider("monitor-collector", &define{})
}
