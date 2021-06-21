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

package metric

import (
	"net/http"

	"github.com/erda-project/erda-infra/modcom/api"
)

func (p *provider) metricQueryHistogram(req *http.Request, params struct {
	Scope string `query:"scope" validate:"required"`
}) interface{} {
	result, err := p.proxy(params.Scope, "histogram", req)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(result)
}

func (p *provider) metricQuery(req *http.Request, params struct {
	Scope string `query:"scope" validate:"required"`
}) interface{} {
	result, err := p.proxy(params.Scope, "query", req)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(result)
}

func (p *provider) metricQueryRange(req *http.Request, params struct {
	Scope string `query:"scope" validate:"required"`
}) interface{} {
	result, err := p.proxy(params.Scope, "range", req)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(result)
}

func (p *provider) metricQueryApdex(req *http.Request, params struct {
	Scope string `query:"scope" validate:"required"`
}) interface{} {
	result, err := p.proxy(params.Scope, "apdex", req)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(result)
}

func (p *provider) metricQueryByQL(req *http.Request) interface{} {
	result, err := p.proxy("", "", req)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(result)
}

func (p *provider) listGroups(req *http.Request, params struct {
	ScopeId string `query:"scopeId" validate:"required"`
}) interface{} {
	result, err := p.proxyBlocks(params.ScopeId, "", req)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(result)
}

func (p *provider) getGroup(req *http.Request, params struct {
	Id      string `query:"id" validate:"required"`
	ScopeId string `query:"scopeId" validate:"required"`
}) interface{} {
	result, err := p.proxyBlocks(params.ScopeId, params.Id, req)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(result)
}
