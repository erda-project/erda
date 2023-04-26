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

package log_http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/filter"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	Name = "log-http"
)

var (
	_ filter.Filter = (*LogHttp)(nil)
)

func init() {
	filter.Register(Name, New)
}

type LogHttp struct{}

func New(_ json.RawMessage) (filter.Filter, error) {
	return &LogHttp{}, nil
}

func (f *LogHttp) OnHttpRequest(ctx context.Context, _ http.ResponseWriter, r *http.Request) (filter.Signal, error) {
	var l = ctx.Value(filter.LoggerCtxKey{}).(logs.Logger)
	var m = map[string]any{
		"scheme":     r.URL.Scheme,
		"host":       r.Host,
		"uri":        r.RequestURI,
		"headers":    r.Header,
		"remoteAddr": r.RemoteAddr,
	}
	if strutil.Equal(r.Header.Get("Content-Type"), "application/json", true) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			l.Errorf(`[LogHttp] failed to io.ReadAll(r.Body), err: %v`, err)
		} else {
			defer func() {
				r.Body = io.NopCloser(bytes.NewReader(data))
			}()
			m["body"] = json.RawMessage(data)
		}
	}
	l.Infof("[LogHttp] request info:\n%s\n", strutil.TryGetYamlStr(m))
	return filter.Continue, nil
}

func (f *LogHttp) OnHttpResponse(ctx context.Context, response *http.Response) (filter.Signal, error) {
	var l = ctx.Value(filter.LoggerCtxKey{}).(logs.Logger)
	var m = map[string]any{
		"headers": response.Header,
	}
	if strutil.Equal(response.Header.Get("Content-Type"), "application/json", true) {
		data, err := io.ReadAll(response.Body)
		if err != nil {
			l.Errorf(`[LogHttp] failed to io.ReadAll(r.Body), err: %v`, err)
		} else {
			defer func() {
				response.Body = io.NopCloser(bytes.NewReader(data))
			}()
			m["body"] = json.RawMessage(data)
		}
	}
	l.Infof("[LogHttp] request info:\n%s\n", strutil.TryGetYamlStr(m))
	return filter.Continue, nil
}
