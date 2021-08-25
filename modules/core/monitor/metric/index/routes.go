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
	"net/http"
	"strings"

	"github.com/erda-project/erda-infra/providers/httpserver"
	api "github.com/erda-project/erda/pkg/common/httpapi"
)

func (p *provider) intRoutes(routes httpserver.Router) error {
	// metric query apis
	routes.GET("/api/metrics-index-manager/inspect/indices", p.getIndicesCache)
	routes.GET("/api/metrics-index-manager/inspect/ttl", p.getTTLConfigCache)
	routes.GET("/api/metrics-index-manager/inspect/ttl-keys", p.getTTLKeysCache)
	routes.GET("/api/metrics-index-manager/inspect/created", p.getCreatedCache)
	routes.GET("/api/metrics-index-manager/inspect/merge", p.getMergeIndices)
	routes.POST("/api/metrics-index-manager/inspect/merge", p.doIndicesMerge)
	return nil
}

func (p *provider) getIndicesCache(r *http.Request) interface{} {
	v := p.m.indices.Load()
	return api.Success(v)
}

func (p *provider) getTTLConfigCache(r *http.Request) interface{} {
	v := p.m.iconfig.Load()
	if v != nil {
		mc := v.(*metricConfig)
		return mc.matcher.SprintTree(true)
	}
	return ""
}

func (p *provider) getTTLKeysCache(r *http.Request) interface{} {
	v := p.m.iconfig.Load()
	if v != nil {
		mc := v.(*metricConfig)
		return api.Success(mc.keysTTL)
	}
	return api.Success(nil)
}

func (p *provider) getCreatedCache(r *http.Request) interface{} {
	p.m.createdLock.Lock()
	defer p.m.createdLock.Unlock()
	created := make(map[string]bool)
	for k, v := range p.m.created {
		created[k] = v
	}
	return api.Success(created)
}

func (p *provider) getMergeIndices(params struct {
	Metric string `query:"metric"`
	Size   string `query:"size" validate:"required"`
}) interface{} {
	params.Metric = normalizeIndexPart(strings.ToLower(params.Metric))
	merges, _, err := p.m.MergeIndices(func(index *IndexEntry) bool {
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

func (p *provider) doIndicesMerge(params struct {
	Metric string `query:"metric"`
	Delete bool   `query:"delete"`
	Size   string `query:"size" validate:"required"`
}) interface{} {
	params.Metric = normalizeIndexPart(strings.ToLower(params.Metric))
	merges, resps, err := p.m.MergeIndices(func(index *IndexEntry) bool {
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
