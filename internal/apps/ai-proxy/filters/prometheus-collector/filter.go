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
	"net/http"
	"strconv"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/metrics"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/provider"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	Name = "prometheus-collector"
)

var (
	_ reverseproxy.RequestFilter  = (*PrometheusCollector)(nil)
	_ reverseproxy.ResponseFilter = (*PrometheusCollector)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

func New(_ json.RawMessage) (reverseproxy.Filter, error) {
	return &PrometheusCollector{DefaultResponseFilter: reverseproxy.NewDefaultResponseFilter()}, nil
}

type PrometheusCollector struct {
	*reverseproxy.DefaultResponseFilter

	labels metrics.LabelValues
}

func (f *PrometheusCollector) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	f.labels.ChatType = infor.Header().Get("X-Erda-AI-Proxy-ChatType")
	f.labels.ChatTitle = infor.Header().Get("X-Erda-AI-Proxy-ChatTitle")
	f.labels.Source = infor.Header().Get("X-Erda-AI-Proxy-Source")
	f.labels.UserId = infor.Header().Get("X-Erda-AI-Proxy-JobNumber")
	f.labels.UserName = infor.Header().Get("X-Erda-AI-Proxy-Name")
	f.labels.Provider = ctx.Value(reverseproxy.ProviderCtxKey{}).(*provider.Provider).Name
	f.labels.Model = f.getModel(ctx, infor)
	f.labels.OperationId = infor.Method()
	if infor.URL() != nil {
		f.labels.OperationId += " " + infor.URL().Path
	}
	return reverseproxy.Continue, nil
}

func (f *PrometheusCollector) OnResponseEOF(ctx context.Context, infor reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) error {
	if err := f.DefaultResponseFilter.OnResponseEOF(ctx, infor, w, chunk); err != nil {
		return err
	}

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
	return nil
}

func (f *PrometheusCollector) getModel(ctx context.Context, infor reverseproxy.HttpInfor) string {
	var l = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger).Sub("PrometheusCollector").Sub("getModel")
	if !httputil.HeaderContains(infor.Header()[httputil.ContentTypeKey], httputil.ApplicationJson) {
		return "-" // todo: Only Content-Type: application/json auditing is supported for now.
	}
	var m = make(map[string]json.RawMessage)
	if err := json.Unmarshal(f.Bytes(), &m); err != nil {
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
