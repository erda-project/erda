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

package org_id

import (
	"net/http"
	"strings"

	"github.com/erda-project/erda-proto-go/common/pb"
	openapiauth "github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth"
	"github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth/over_permission/match"
	"github.com/erda-project/erda/pkg/http/httputil"
)

type overPermissionOrgId struct {
	provider *provider
}

func newOverPermissionOrgId(provider *provider) *overPermissionOrgId {
	return &overPermissionOrgId{provider: provider}
}

func (o *overPermissionOrgId) Weight() int64 {
	return o.provider.Cfg.Weight
}

func (o *overPermissionOrgId) Match(r *http.Request, opts openapiauth.Options) (bool, interface{}) {
	opt := opts.Get(match.ProtoComponent)
	if opt == nil {
		return false, nil
	}
	checkOverPermission := opt.(*pb.CheckOverPermission)
	if checkOverPermission == nil {
		return false, nil
	}
	if checkOverPermission.OrgId == nil || !checkOverPermission.OrgId.Enable {
		return false, nil
	}

	matchExpr := checkOverPermission.OrgId.Expr
	if len(matchExpr) <= 0 {
		return false, nil
	}
	matchExpr = trim(matchExpr)
	m := match.Get(matchExpr, r)
	if len(m) == 0 {
		return false, nil
	}
	return true, m
}

func trim(arr []string) []string {
	var result []string
	for _, i := range arr {
		result = append(result, strings.TrimSpace(i))
	}
	return result
}

func (o *overPermissionOrgId) Check(r *http.Request, data interface{}, opts openapiauth.Options) (bool, *http.Request, error) {
	orgId := r.Header.Get(httputil.OrgHeader)
	if len(orgId) == 0 {
		return false, r, nil
	}
	if d, ok := data.(map[string]interface{}); ok {
		if d["inParams.scope"] == "org" && d["inParams.scopeId"] != orgId {
			return false, r, nil
		}
	}
	return true, r, nil
}
