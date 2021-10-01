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

package indexmanager

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/erda-project/erda-infra/providers/httpserver"
	indexloader "github.com/erda-project/erda/modules/core/monitor/metric/index-loader"
	api "github.com/erda-project/erda/pkg/common/httpapi"
)

func (p *provider) intRoutes(routes httpserver.Router) error {
	routes.GET("/api/metrics-index-manager/inspect/ttl", p.getTTLConfigCache)
	routes.GET("/api/metrics-index-manager/inspect/ttl-keys", p.getTTLKeysCache)
	routes.GET("/api/metrics-index-manager/inspect/created", p.getCreatedCache)
	routes.GET("/api/metrics-index-manager/inspect/merge", p.getMergeIndices)
	routes.POST("/api/metrics-index-manager/inspect/merge", p.mergeIndices)
	routes.POST("/api/metrics-index-manager/inspect/rollover", p.rolloverIndicesByRequest)
	return nil
}

func (p *provider) getTTLConfigCache(r *http.Request) interface{} {
	mc, _ := p.iconfig.Load().(*metricConfig)
	if mc != nil {
		return mc.matcher.SprintTree(true)
	}
	return ""
}

func (p *provider) getTTLKeysCache(r *http.Request) interface{} {
	mc, _ := p.iconfig.Load().(*metricConfig)
	if mc != nil {
		return api.Success(mc.keysTTL)
	}
	return api.Success(nil)
}

func (p *provider) getCreatedCache(r *http.Request) interface{} {
	p.createdLock.Lock()
	defer p.createdLock.Unlock()
	created := make(map[string]bool)
	for k, v := range p.created {
		created[k] = v
	}
	return api.Success(created)
}

func (p *provider) getMergeIndices(params struct {
	Metric string `query:"metric"`
	Size   string `query:"size" validate:"required"`
}) interface{} {
	params.Metric = normalizeIndexSegmentName(strings.ToLower(params.Metric))
	merges, _, err := p.MergeIndices(context.Background(), func(index *indexloader.IndexEntry) bool {
		if len(params.Metric) > 0 {
			return index.Metric == params.Metric
		}
		return true
	}, params.Size, false, false)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(merges)
}

func (p *provider) mergeIndices(params struct {
	Metric string `query:"metric"`
	Delete bool   `query:"delete"`
	Size   string `query:"size" validate:"required"`
}) interface{} {
	params.Metric = normalizeIndexSegmentName(strings.ToLower(params.Metric))
	merges, resps, err := p.MergeIndices(context.Background(), func(index *indexloader.IndexEntry) bool {
		if len(params.Metric) > 0 {
			return index.Metric == params.Metric
		}
		return true
	}, params.Size, true, params.Delete)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(map[string]interface{}{
		"merges":    merges,
		"responses": resps,
	})
}

func (p *provider) rolloverIndicesByRequest(params struct {
	Metric     string `query:"metric" validate:"required"`
	Namespacce string `query:"namespacce" validate:"required"`
	Size       string `query:"size" validate:"required"`
}) interface{} {
	params.Metric = normalizeIndexSegmentName(strings.ToLower(params.Metric))
	params.Namespacce = normalizeIndexSegmentName(strings.ToLower(params.Namespacce))
	alias := p.indexAlias(params.Metric, params.Namespacce)
	body, _ := json.Marshal(map[string]interface{}{
		"conditions": map[string]interface{}{
			"max_size": params.Size,
		},
	})
	ok, err := p.rolloverAlias(alias, string(body))
	if err != nil {
		return api.Errors.Internal(err)
	}
	if ok {
		p.Loader.ReloadIndices()
	}
	return api.Success(map[string]interface{}{
		"ok": ok,
	})
}
