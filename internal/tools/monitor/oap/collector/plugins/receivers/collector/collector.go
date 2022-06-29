// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/labstack/echo"

	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib/protoparser/rawparser"
)

const (
	headerCollectorFormat       = "x-erda-collector-format"
	headerContentEncoding       = "Content-Encoding"
	headerCustomContentEncoding = "Custom-Content-Encoding"
)

// Collect internal logs from outside
func (p *provider) collectLogs(ctx echo.Context) error {
	source := ctx.Param("source")
	if source == "" {
		return ctx.NoContent(http.StatusBadRequest)
	}
	name := source + "_log"
	req := ctx.Request()

	err := rawparser.ParseStream(
		req.Body,
		req.Header.Get(headerContentEncoding),
		req.Header.Get(headerCustomContentEncoding),
		req.Header.Get(headerCollectorFormat), func(buf []byte) error {
			return p.sendRaw(name, buf)
		})
	if err != nil {
		return fmt.Errorf("parse stream: %w", err)
	}
	return ctx.NoContent(http.StatusNoContent)
}

// Collect internal metrics from outside
func (p *provider) collectMetric(ctx echo.Context) error {
	contentType := ctx.Request().Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		return p.parseJSON(ctx, ctx.Param("metric"))
	}
	return p.parseLine(ctx, ctx.Param("metric"))
}

func (p *provider) parseJSON(ctx echo.Context, name string) error {
	body, err := lib.ReadBody(ctx.Request())
	if err != nil {
		return fmt.Errorf("parseJSON readBody: %w", err)
	}
	if _, err := jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		if err := p.sendRaw(name, value); err != nil {
			p.Log.Errorf("sendRaw err: %s", err)
		}
	}, name); err != nil {
		return fmt.Errorf("parseJSON ArrayEach: %w", err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (p *provider) parseLine(ctx echo.Context, name string) error {
	r, err := lib.ReadBody(ctx.Request())
	if err != nil {
		return fmt.Errorf("parseLine readBody: %w", err)
	}
	buf := bufio.NewReader(bytes.NewReader(r))
	for {
		line, err := buf.ReadString('\n')
		if e := p.sendRaw(name, []byte(line)); e != nil {
			return e
		}
		if err != nil || err == io.EOF {
			break
		}
	}
	return ctx.NoContent(http.StatusNoContent)
}
