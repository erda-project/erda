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

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httputil"
)

// GetApp get app by id from cmdb.
func (b *Bundle) GetApp(id uint64) (*apistructs.ApplicationDTO, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var fetchResp apistructs.ApplicationFetchResponse
	resp, err := hc.Get(host).Path(fmt.Sprintf("/api/applications/%d", id)).Header(httputil.InternalHeader, "bundle").Do().JSON(&fetchResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !fetchResp.Success {
		return nil, toAPIError(resp.StatusCode(), fetchResp.Error)
	}

	return &fetchResp.Data, nil
}

func (b *Bundle) GetMyApps(userid string, orgid uint64) (*apistructs.ApplicationListResponseData, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var listResp apistructs.ApplicationListResponse
	resp, err := hc.Get(host).
		Path("/api/applications/actions/list-my-applications").
		Header(httputil.OrgHeader, strconv.FormatUint(orgid, 10)).
		Header(httputil.UserHeader, userid).
		Param("pageSize", "9999").
		Param("pageNo", "1").
		Do().JSON(&listResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !listResp.Success {
		return nil, toAPIError(resp.StatusCode(), listResp.Error)
	}
	return &listResp.Data, nil
}

// GetAppsByProject 根据 projectID 获取应用列表
func (b *Bundle) GetAppsByProject(projectID, orgID uint64, userID string) (*apistructs.ApplicationListResponseData, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var listResp apistructs.ApplicationListResponse
	resp, err := hc.Get(host).
		Path(fmt.Sprintf("/api/applications")).
		Header(httputil.OrgHeader, strconv.FormatUint(orgID, 10)).
		Header(httputil.UserHeader, userID).
		Param("projectId", strconv.FormatUint(projectID, 10)).
		Param("pageSize", "100").
		Param("pageNo", "1").
		Do().JSON(&listResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !listResp.Success {
		return nil, toAPIError(resp.StatusCode(), listResp.Error)
	}

	return &listResp.Data, nil
}

// get applications by projectID and app name
func (b *Bundle) GetAppsByProjectAndAppName(projectID, orgID uint64, userID string, appName string) (*apistructs.ApplicationListResponseData, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var listResp apistructs.ApplicationListResponse
	resp, err := hc.Get(host).
		Path("/api/applications").
		Header(httputil.OrgHeader, strconv.FormatUint(orgID, 10)).
		Header(httputil.UserHeader, userID).
		Param("projectId", strconv.FormatUint(projectID, 10)).
		Param("pageSize", "1").
		Param("pageNo", "1").
		Param("name", appName).
		Do().JSON(&listResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !listResp.Success {
		return nil, toAPIError(resp.StatusCode(), listResp.Error)
	}

	return &listResp.Data, nil
}

// GetAppsByProjectSimple 根据 projectID 获取应用列表简单信息
func (b *Bundle) GetAppsByProjectSimple(projectID, orgID uint64, userID string) (*apistructs.ApplicationListResponseData, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var listResp apistructs.ApplicationListResponse
	resp, err := hc.Get(host).
		Path(fmt.Sprintf("/api/applications")).
		Header(httputil.OrgHeader, strconv.FormatUint(orgID, 10)).
		Header(httputil.UserHeader, userID).
		Param("projectId", strconv.FormatUint(projectID, 10)).
		Param("pageSize", "100").
		Param("pageNo", "1").
		Param("isSimple", "true").
		Do().JSON(&listResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !listResp.Success {
		return nil, toAPIError(resp.StatusCode(), listResp.Error)
	}

	return &listResp.Data, nil
}

// Deprecated
type AbilityAppReq struct {
	OrgId           uint64 `json:"orgId"`
	ClusterId       uint64 `json:"clusterId"`
	ClusterName     string `json:"clusterName"`
	ApplicationName string `json:"applicationName"`
	Operator        string `json:"operator"`
}

// GetAppPublishItemRelationsGroupByENV 根据 appID 获取应用关联的发布内容
func (b *Bundle) GetAppPublishItemRelationsGroupByENV(appID uint64) (*apistructs.QueryAppPublishItemRelationGroupByENVResponse, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var relationResp apistructs.QueryAppPublishItemRelationGroupByENVResponse
	resp, err := hc.Get(host).
		Path(fmt.Sprintf("/api/applications/%d/actions/get-publish-item-relations", appID)).
		Do().JSON(&relationResp)
	if err != nil {
		return nil, err
	}
	if !resp.IsOK() {
		return nil, errors.Errorf("failed to fetch publish item relation, statusCode: %d", resp.StatusCode())
	}
	if !relationResp.Success {
		return nil, errors.Errorf("failed to fetch publish item relation, err: %s", relationResp.Error.Msg)
	}
	return &relationResp, nil
}

// QueryAppPublishItemRelations 查询应用关联的发布内容
func (b *Bundle) QueryAppPublishItemRelations(req *apistructs.QueryAppPublishItemRelationRequest) (*apistructs.QueryAppPublishItemRelationResponse, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var relationResp apistructs.QueryAppPublishItemRelationResponse
	resp, err := hc.Get(host).
		Path("/api/applications/actions/query-publish-item-relations").
		Param("publishItemID", strconv.FormatInt(req.PublishItemID, 10)).
		Param("appID", strconv.FormatInt(req.AppID, 10)).
		Param("ak", req.AK).Param("ai", req.AI).
		Do().JSON(&relationResp)
	if err != nil {
		return nil, err
	}
	if !resp.IsOK() {
		return nil, errors.Errorf("failed to fetch publish item relation, statusCode: %d", resp.StatusCode())
	}
	if !relationResp.Success {
		return nil, errors.Errorf("failed to fetch publish item relation, err: %s", relationResp.Error.Msg)
	}
	return &relationResp, nil
}

func (b *Bundle) RemoveAppPublishItemRelations(publishItemID int64) error {
	host, err := b.urls.CMDB()
	if err != nil {
		return err
	}
	req := apistructs.RemoveAppPublishItemRelationsRequest{
		PublishItemId: publishItemID,
	}
	hc := b.hc
	var getResp apistructs.RemoveAppPublishItemRelationsResponse
	resp, err := hc.Post(host).Path("/api/applications/actions/remove-publish-item-relations").
		Header("Internal-Client", "bundle").JSONBody(req).
		Do().JSON(&getResp)
	if err != nil {
		apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !getResp.Success {
		return toAPIError(resp.StatusCode(), getResp.Error)
	}
	return nil
}

// get my apps by paging
func (b *Bundle) GetAllMyApps(userid string, orgid uint64, req apistructs.ApplicationListRequest) (*apistructs.ApplicationListResponseData, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var listResp apistructs.ApplicationListResponse
	resp, err := hc.Get(host).
		Path("/api/applications/actions/list-my-applications").
		Header(httputil.OrgHeader, strconv.FormatUint(orgid, 10)).
		Header(httputil.UserHeader, userid).
		Param("pageSize", strconv.Itoa(req.PageSize)).
		Param("pageNo", strconv.Itoa(req.PageNo)).
		Param("name", req.Name).
		Param("mode", req.Mode).
		Param("q", req.Query).
		Param("public", req.Public).
		Param("projectId", strconv.FormatUint(req.ProjectID, 10)).
		Param("isSimple", strconv.FormatBool(req.IsSimple)).
		Param("orderBy", req.OrderBy).
		Do().JSON(&listResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !listResp.Success {
		return nil, toAPIError(resp.StatusCode(), listResp.Error)
	}
	return &listResp.Data, nil
}
