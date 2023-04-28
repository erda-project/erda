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

package prometheus_collector

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/filter"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/metrics"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/provider"
	"github.com/erda-project/erda/pkg/http/httputil"
)

const (
	Name = "prometheus-collector"
)

var (
	_ filter.RequestGetterFilter  = (*PrometheusCollector)(nil)
	_ filter.ResponseGetterFilter = (*PrometheusCollector)(nil)
)

func init() {
	filter.Register(Name, New)
}

type PrometheusCollector struct {
	labels metrics.LabelValues
}

func New(_ json.RawMessage) (filter.Filter, error) {
	return &PrometheusCollector{}, nil
}

func (f *PrometheusCollector) OnHttpRequestGetter(ctx context.Context, infor filter.HttpInfor) (filter.Signal, error) {
	f.labels.ChatType = infor.Header().Get("X-Erda-AI-Proxy-ChatType")
	f.labels.ChatTitle = infor.Header().Get("X-Erda-AI-Proxy-ChatTitle")
	f.labels.Source = infor.Header().Get("X-Erda-AI-Proxy-Source")
	f.labels.UserId = infor.Header().Get("X-Erda-AI-Proxy-JobNumber")
	f.labels.UserName = infor.Header().Get("X-Erda-AI-Proxy-Name")
	f.labels.Provider = ctx.Value(filter.ProviderCtxKey{}).(*provider.Provider).Name
	f.labels.Model = f.getModel(ctx, infor)
	f.labels.OperationId = infor.Method()
	if infor.URL() != nil {
		f.labels.OperationId += " " + infor.URL().Path
	}
	return filter.Continue, nil
}

func (f *PrometheusCollector) OnHttpResponseGetter(_ context.Context, infor filter.HttpInfor) (filter.Signal, error) {
	f.labels.Status = infor.Status()
	f.labels.StatusCode = strconv.FormatInt(int64(infor.StatusCode()), 10)
	for _, v := range []*string{
		&f.labels.ChatTitle,
		&f.labels.ChatType,
		&f.labels.UserId,
		&f.labels.UserName,
	} {
		if data, err := base64.StdEncoding.DecodeString(*v); err == nil {
			*v = string(data)
		}
	}
	metrics.Get().WithLabelValues(f.labels.Values()...).Inc()
	return filter.Continue, nil
}

func (f *PrometheusCollector) getModel(ctx context.Context, infor filter.HttpInfor) string {
	var l = ctx.Value(filter.LoggerCtxKey{}).(logs.Logger).Sub("PrometheusCollector").Sub("getModel")
	if !httputil.HeaderContains(infor.Header()[httputil.ContentTypeKey], httputil.ApplicationJson) {
		return "-" // todo: Only Content-Type: application/json auditing is supported for now.
	}
	body, err := infor.Body()
	if err != nil {
		l.Errorf("failed to infor.Body, err: %v")
		return "-"
	}
	var m = make(map[string]json.RawMessage)
	if err := json.NewDecoder(body).Decode(&m); err != nil {
		l.Errorf("failed to Decode r.Body to m (%T), err: %v", m, err)
		return "-"
	}
	data, ok := m["model"]
	if !ok {
		l.Debug(`no field "model" in r.Body`)
		return "-"
	}
	var model string
	if err := json.Unmarshal(data, &model); err != nil {
		l.Errorf("failed to json.Unmarshal %s to string, err: %v", string(data), err)
		return "-"
	}
	return model
}
