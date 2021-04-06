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

package endpoints

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/strutil"
)

// ListDomains 查询域名列表
func (e *Endpoints) ListDomains(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	v := vars["runtimeID"]
	runtimeID, err := strutil.Atoi64(v)
	if err != nil {
		return apierrors.ErrListDomain.InvalidParameter(strutil.Concat("runtimeID: ", v)).ToResp(), nil
	}
	orgID, err := getOrgID(r)
	if err != nil {
		return apierrors.ErrListDomain.InvalidParameter(err).ToResp(), nil
	}
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrListDomain.NotLogin().ToResp(), nil
	}
	data, err := e.domain.List(userID, orgID, uint64(runtimeID))
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(data)
}

// UpdateDomains 更新域名
func (e *Endpoints) UpdateDomains(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	v := vars["runtimeID"]
	runtimeID, err := strutil.Atoi64(v)
	if err != nil {
		return apierrors.ErrUpdateDomain.InvalidParameter(strutil.Concat("runtimeID: ", v)).ToResp(), nil
	}
	orgID, err := getOrgID(r)
	if err != nil {
		return apierrors.ErrUpdateDomain.InvalidParameter(err).ToResp(), nil
	}
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrUpdateDomain.NotLogin().ToResp(), nil
	}
	var group apistructs.DomainGroup
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		return apierrors.ErrUpdateDomain.InvalidParameter(err).ToResp(), nil
	}
	if err := e.domain.Update(userID, orgID, uint64(runtimeID), &group); err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(nil)
}
