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

package over_permission

import (
	"net/http"

	"github.com/erda-project/erda-proto-go/common/pb"
	openapiauth "github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth"
	"github.com/erda-project/erda/internal/core/openapi/openapi-ng/auth/over_permission/match"
)

const compent = "OverPermission"

type overPermissionOrg struct {
	provider *provider
}

func newOverPermissionOrg(p *provider) *overPermissionOrg {
	return &overPermissionOrg{
		p,
	}
}

func (o *overPermissionOrg) Weight() int64 {
	return o.provider.Cfg.Weight
}

func (o *overPermissionOrg) Match(r *http.Request, opts openapiauth.Options) (bool, interface{}) {
	opt := opts.Get(compent)
	if opt == nil {
		return false, nil
	}
	checkOverPermission := opt.(*pb.CheckOverPermission)
	if checkOverPermission == nil {
		return false, nil
	}
	if checkOverPermission.Org == nil || !checkOverPermission.Org.Enable {
		return false, nil
	}

	matchExpr := checkOverPermission.Org.Expr
	if len(matchExpr) <= 0 {
		matchExpr = o.provider.Cfg.DefaultMatchOrg
	}

	m := match.Get(matchExpr, r)
	if len(m) == 0 {
		return false, nil
	}
	return true, m
}

func (o *overPermissionOrg) Check(r *http.Request, data interface{}, opts openapiauth.Options) (bool, *http.Request, error) {
	org := r.Header.Get("org")
	if len(org) == 0 {
		o.provider.Log.Debug("org name checker, header org should be not nil")
		return false, r, nil
	}
	if d, ok := data.(map[string]interface{}); ok {
		if d["scope"] == "org" && d["scopeId"] != org {
			return false, r, nil
		}
	}
	return true, r, nil
}
