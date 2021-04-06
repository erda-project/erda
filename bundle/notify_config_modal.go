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

package bundle

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/httputil"
)

func (b *Bundle) CreateOrEditNotify(submitData *apistructs.EditOrCreateModalData, inParams *apistructs.InParams, userId string) error {
	host, err := b.urls.Monitor()
	if err != nil {
		return err
	}
	hc := b.hc
	var resp apistructs.Header
	var path string
	createMap := make(map[string]interface{})
	createMap["templateId"] = submitData.Items
	createMap["notifyName"] = submitData.Name
	createMap["notifyGroupId"] = submitData.Target
	createMap["channels"] = submitData.Channels
	var httpResp *httpclient.Response
	if submitData.Id != 0 {
		path = fmt.Sprintf("/api/notify/records/%d?scope=%v&scopeId=%v", submitData.Id, inParams.ScopeType, inParams.ScopeId)
		httpResp, err = hc.Put(host).Path(path).JSONBody(createMap).Header(httputil.UserHeader, userId).Do().JSON(&resp)
	} else {
		path = fmt.Sprintf("/api/notify/records?scope=%v&scopeId=%v", inParams.ScopeType, inParams.ScopeId)
		httpResp, err = hc.Post(host).Path(path).JSONBody(createMap).Header(httputil.UserHeader, userId).Do().JSON(&resp)
	}
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !resp.Success {
		return toAPIError(httpResp.StatusCode(), resp.Error)
	}
	return nil
}
