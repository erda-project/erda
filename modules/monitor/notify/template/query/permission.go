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

package query

import (
	"fmt"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/modules/monitor/common/permission"
)

func (p *provider) getScope(ctx httpserver.Context) (string, error) {
	req := ctx.Request()
	scope := req.URL.Query().Get("scope")
	if scope != "" {
		if scope == permission.ScopeMicroService {
			return permission.ScopeProject, nil
		}
		return scope, nil
	}
	err := fmt.Errorf("the scope must not empty")
	return "", err
}

func (p *provider) getScopeID(ctx httpserver.Context) (string, error) {
	var id string
	req := ctx.Request()
	scopeId := req.URL.Query().Get("scopeId")
	if scopeId != "" {
		id = scopeId
	} else {
		err := fmt.Errorf("the scopeID must not empty")
		return "", err
	}
	scope, err := p.getScope(ctx)
	if err != nil {
		return "", err
	}
	if scope != permission.ScopeProject {
		return id, nil
	} else {
		scopeID, err := p.N.GetProjectByScopeID(id)
		if err != nil {
			return "", err
		}
		return scopeID, nil
	}
}
