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
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

// GetNotifyGroupDetail 查询通知组详情
func (b *Bundle) GetNotifyGroupDetail(id int64, orgID int64, userID string) (*apistructs.NotifyGroupDetail, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var getResp apistructs.GetNotifyGroupDetailResponse
	resp, err := hc.Get(host).Path(strutil.Concat("/api/notify-groups/", strconv.FormatInt(id, 10), "/detail")).
		Header("Org-ID", strconv.FormatInt(orgID, 10)).
		Header("User-ID", userID).
		Do().JSON(&getResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !getResp.Success {
		return nil, toAPIError(resp.StatusCode(), getResp.Error)
	}
	return &getResp.Data, nil
}

func (b *Bundle) QueryNotifiesBySource(orgID string, sourceType, sourceID, itemName, label string, clusterNames ...string) ([]*apistructs.NotifyDetail, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	clusterName := ""
	if len(clusterNames) > 0 {
		clusterName = clusterNames[0]
	}
	var getResp apistructs.QuerySourceNotifyResponse
	resp, err := hc.Get(host).Path("/api/notifies/actions/search-by-source").
		Param("sourceType", sourceType).
		Param("sourceId", sourceID).
		Param("itemName", itemName).
		Param("orgId", orgID).
		Param("clusterName", clusterName).
		Param("label", label).
		Do().JSON(&getResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !getResp.Success {
		return nil, toAPIError(resp.StatusCode(), getResp.Error)
	}
	return getResp.Data, nil
}

func (b *Bundle) CreateNotifyHistory(request *apistructs.CreateNotifyHistoryRequest) (int64, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return 0, err
	}
	hc := b.hc

	var getResp apistructs.CreateNotifyHistoryResponse
	resp, err := hc.Post(host).Path("/api/notify-histories").JSONBody(request).
		Do().JSON(&getResp)
	if err != nil {
		return 0, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !getResp.Success {
		return 0, toAPIError(resp.StatusCode(), getResp.Error)
	}
	return getResp.Data, nil
}

// GetNotifyConfig 获取通知配置
func (b *Bundle) GetNotifyConfig(orgIDstr, userID string) (*apistructs.NotifyConfigUpdateRequestBody, error) {
	// TODO: userID should be deprecated
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var resp apistructs.NotifyConfigGetResponse
	r, err := hc.Get(host).Path(fmt.Sprintf("/api/orgs/%s/actions/get-notify-config", orgIDstr)).
		Header(httputil.InternalHeader, "bundle").
		Header("User-ID", userID). // TODO: for compatibility
		Do().
		JSON(&resp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() || !resp.Success {
		return nil, toAPIError(r.StatusCode(), resp.Error)
	}
	return &resp.Data, nil
}
