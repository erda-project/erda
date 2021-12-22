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

package promremotewrite

import (
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	"github.com/erda-project/erda/modules/oap/collector/plugins"
	"github.com/prometheus/prometheus/prompb"
)

var providerName = plugins.WithPrefixReceiver("prometheus-remote-write")

type config struct {
}

// +provider
type provider struct {
	Cfg    *config
	Log    logs.Logger
	Router httpserver.Router `autowired:"http-router"`

	label         string
	consumerFuncs []model.MetricReceiverConsumeFunc
	mu            sync.RWMutex
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	p.label = ctx.Label()
	p.consumerFuncs = make([]model.MetricReceiverConsumeFunc, 0)
	p.Router.POST("/api/v1/prometheus-remote-write", p.prwHandler)
	return nil
}

func (p *provider) prwHandler(req *http.Request, resp http.ResponseWriter) {
	defer func() {
		_ = req.Body.Close()
	}()

	buf, err := io.ReadAll(req.Body)
	if err != nil {
		p.Log.Errorf("read body error: %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	wr := &prompb.WriteRequest{}
	err = wr.Unmarshal(buf)
	if err != nil {
		p.Log.Errorf("unmarshal body error: %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	ms := convertToMetrics(wr)

	p.mu.RLock()
	for _, fn := range p.consumerFuncs {
		fn(ms.Clone())
	}
	p.mu.RUnlock()

	resp.WriteHeader(http.StatusNoContent)
}

func convertToMetrics(wr *prompb.WriteRequest) model.Metrics {
	return model.Metrics{}
}

func (p *provider) RegisterConsumeFunc(consumer model.MetricReceiverConsumeFunc) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.consumerFuncs = append(p.consumerFuncs, consumer)
}

func (p *provider) ComponentID() model.ComponentID {
	return model.ComponentID(strings.Join([]string{providerName, p.label}, "@"))
}

func init() {
	servicehub.Register(providerName, &servicehub.Spec{
		Services:    []string{},
		Description: "here is description of prometheus-remote-write",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
