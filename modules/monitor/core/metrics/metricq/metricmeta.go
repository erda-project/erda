// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package metricq

import (
	"net/http"
	"strings"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq/metricmeta"
)

func (p *provider) listMetricNames(r *http.Request, params struct {
	Scope   string `query:"scope" validate:"required"`
	ScopeID string `query:"scopeId" validate:"required"`
}) interface{} {
	names, err := p.q.MetricNames(api.Language(r), params.Scope, params.ScopeID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(names)
}

func (p *provider) listMetricMeta(r *http.Request, params struct {
	Scope   string `query:"scope" validate:"required"`
	ScopeID string `query:"scopeId" validate:"required"`
	Metrics string `query:"metrics"`
}) interface{} {
	var names []string
	if len(params.Metrics) > 0 {
		names = strings.Split(params.Metrics, ",")
	}
	metrics, err := p.q.MetricMeta(api.Language(r), params.Scope, params.ScopeID, names...)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(metrics)
}

func (p *provider) listMetricGroups(r *http.Request, params struct {
	Scope   string `query:"scope" validate:"required"`
	ScopeID string `query:"scopeId" validate:"required"`
	Mode    string `query:"mode"`
}) interface{} {
	groups, err := p.q.MetricGroups(api.Language(r), params.Scope, params.ScopeID, params.Mode)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(groups)
}

func (p *provider) getMetricGroup(r *http.Request, params struct {
	Scope      string `query:"scope" validate:"required"`
	ScopeID    string `query:"scopeId" validate:"required"`
	ID         string `param:"id" validate:"required"`
	Mode       string `query:"mode"`
	Version    string `query:"version"`
	Format     string `query:"format"`
	AppendTags bool   `query:"appendTags"`
}) interface{} {
	if len(params.Format) <= 0 {
		if params.Version == "v2" {
			params.Format = metricmeta.InfluxFormat
			params.AppendTags = true
		} else if len(params.Format) <= 0 && params.Mode != "analysis" {
			// 标品大盘需要点格式，但告警表达式不支持点格式，所以非告警模式的元数据查询，都用点格式。
			params.Format = metricmeta.DotFormat
		}
	}
	group, err := p.q.MetricGroup(api.Language(r), params.Scope, params.ScopeID, params.ID, params.Mode, params.Format, params.AppendTags)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(group)
}
