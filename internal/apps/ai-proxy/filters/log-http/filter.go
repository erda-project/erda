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
	"context"
	"encoding/json"
	"os"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/filter"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	Name = "log-http"
)

var (
	_ filter.RequestInforFilter  = (*LogHttp)(nil)
	_ filter.ResponseInforFilter = (*LogHttp)(nil)
)

func init() {
	filter.Register(Name, New)
}

type LogHttp struct{}

func New(_ json.RawMessage) (filter.Filter, error) {
	return &LogHttp{}, nil
}

func (f *LogHttp) OnHttpRequestInfor(ctx context.Context, infor filter.HttpInfor) (filter.Signal, error) {
	if !strutil.Equal(os.Getenv("LOG_LEVEL"), "debug") {
		return filter.Continue, nil
	}
	var l = ctx.Value(filter.LoggerCtxKey{}).(logs.Logger).Sub("LogHttp")
	var url = infor.URL()
	var m = map[string]any{
		"scheme":     url.Scheme,
		"host":       infor.Host(),
		"url.host":   url.Host,
		"uri":        url.RequestURI(),
		"headers":    infor.Header(),
		"remoteAddr": infor.RemoteAddr(),
	}
	defer func() { l.Debugf("request info: %s", strutil.TryGetJsonStr(m)) }()
	if !httputil.HeaderContains(infor.Header()[httputil.ContentTypeKey], httputil.ApplicationJson) ||
		infor.ContentLength() == 0 {
		return filter.Continue, nil
	}
	body, err := infor.Body()
	if err != nil {
		return filter.Intercept, err
	}
	m["body"] = body.String()
	return filter.Continue, nil
}

func (f *LogHttp) OnHttpResponseInfor(ctx context.Context, infor filter.HttpInfor) (filter.Signal, error) {
	if strutil.Equal(os.Getenv("LOG_LEVEL"), "debug") {
		return filter.Continue, nil
	}
	var l = ctx.Value(filter.LoggerCtxKey{}).(logs.Logger).Sub("LogHttp").Sub("OnHttpResponseInfor")
	var m = map[string]any{
		"headers":     infor.Header(),
		"status":      infor.Status(),
		"status code": infor.StatusCode(),
	}
	defer func() { l.Debugf("response info: %s", strutil.TryGetJsonStr(m)) }()
	if !httputil.HeaderContains(infor.Header()[httputil.ContentTypeKey], httputil.ApplicationJson) {
		return filter.Continue, nil
	}
	body, err := infor.Body()
	if err != nil {
		l.Errorf("failed to infor.Body(), err: %v", err)
		return filter.Intercept, err
	}
	m["body"] = body.String()
	return filter.Continue, nil
}
