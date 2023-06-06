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
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/metrics"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
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

	lvs metrics.LabelValues
}

func (f *PrometheusCollector) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	f.lvs.ChatType = infor.Header().Get(vars.XErdaAIProxyChatType)
	f.lvs.ChatTitle = infor.Header().Get(vars.XErdaAIProxyChatTitle)
	f.lvs.Source = infor.Header().Get(vars.XErdaAIProxySource)
	f.lvs.UserId = infor.Header().Get(vars.XErdaAIProxyJobNumber)
	f.lvs.UserName = infor.Header().Get(vars.XErdaAIProxyName)
	f.lvs.Provider = ctx.Value(vars.CtxKeyProvider{}).(*provider.Provider).Name
	f.lvs.Model = f.getModel(ctx, infor)
	f.lvs.OperationId = infor.Method()
	if infor.URL() != nil {
		f.lvs.OperationId += " " + infor.URL().Path
	}
	return reverseproxy.Continue, nil
}

func (f *PrometheusCollector) OnResponseEOF(ctx context.Context, infor reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) error {
	if err := f.DefaultResponseFilter.OnResponseEOF(ctx, infor, w, chunk); err != nil {
		return err
	}

	f.lvs.Status = infor.Status()
	f.lvs.StatusCode = strconv.FormatInt(int64(infor.StatusCode()), 10)
	for _, v := range []*string{
		&f.lvs.ChatTitle,
		&f.lvs.ChatType,
		&f.lvs.UserId,
		&f.lvs.UserName,
	} {
		if data, err := base64.StdEncoding.DecodeString(*v); err == nil {
			*v = string(data)
		}
	}
	metrics.CounterVec().WithLabelValues(f.lvs.Values()...).Inc()

	return nil
}

func (f *PrometheusCollector) getModel(ctx context.Context, infor reverseproxy.HttpInfor) string {
	var l = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger).Sub("getModel")
	if !httputil.HeaderContains(infor.Header()[httputil.ContentTypeKey], httputil.ApplicationJson) {
		return "-" // todo: Only Content-Type: application/json auditing is supported for now.
	}
	var m = make(map[string]json.RawMessage)
	if body := infor.BodyBuffer(); body != nil {
		if err := json.NewDecoder(body).Decode(&m); err != nil {
			l.Errorf("failed to Decode r.Body to m (%T), err: %v", m, err)
			return "-"
		}
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
