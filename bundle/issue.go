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
	"bytes"
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httputil"
)

func (b *Bundle) PageIssues(req apistructs.IssuePagingRequest) (*apistructs.IssuePagingResponse, error) {
	host, err := b.urls.DOP()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var rsp apistructs.IssuePagingResponse
	httpResp, err := hc.Get(host).Path(fmt.Sprintf("/api/issues")).
		Header(httputil.UserHeader, req.UserID).
		Params(req.UrlQueryString()).
		Do().JSON(&rsp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !rsp.Success {
		return nil, toAPIError(httpResp.StatusCode(), rsp.Error)
	}

	return &rsp, nil
}

// https://terminus-test-org.test.terminus.io/api/labels?type=issue&projectID=1&pageNo=1&pageSize=300
func (b *Bundle) Labels(tp string, projectID uint64, userID string) (*apistructs.ProjectLabelListResponseData, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var rsp apistructs.ProjectLabelListResponse
	httpResp, err := hc.Get(host).Path(fmt.Sprintf("/api/labels")).
		Header(httputil.UserHeader, userID).
		Param("type", tp).
		Param("projectID", strconv.FormatUint(projectID, 10)).
		Param("pageNo", "1").
		Param("pageSize", "300").
		Do().JSON(&rsp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !rsp.Success {
		return nil, toAPIError(httpResp.StatusCode(), rsp.Error)
	}

	return rsp.Data, nil
}

// GetIssue 通过id获取事件
func (b *Bundle) GetIssue(id uint64) (*apistructs.Issue, error) {
	host, err := b.urls.DOP()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var issueResp apistructs.IssueGetResponse
	resp, err := hc.Get(host).Path(fmt.Sprintf("/api/issues/%d", id)).
		Header(httputil.InternalHeader, "bundle").Do().JSON(&issueResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !issueResp.Success {
		return nil, toAPIError(resp.StatusCode(), issueResp.Error)
	}

	return issueResp.Data, nil
}

// CreateIssueTicket 创建工单
// TODO 和ps_ticket的bundle同名了，待前者废弃后改回
func (b *Bundle) CreateIssueTicket(createReq apistructs.IssueCreateRequest) (uint64, error) {
	host, err := b.urls.DOP()
	if err != nil {
		return 0, err
	}
	hc := b.hc

	if createReq.Type != apistructs.IssueTypeTicket {
		return 0, apierrors.ErrInvoke.InvalidParameter("issue type can only be TICKET")
	}

	var createResp apistructs.IssueCreateResponse
	resp, err := hc.Post(host).Path("/api/issues").Header(httputil.InternalHeader, "bundle").
		JSONBody(&createReq).Do().JSON(&createResp)
	if err != nil {
		return 0, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !createResp.Success {
		return 0, toAPIError(resp.StatusCode(), createResp.Error)
	}

	return createResp.Data, nil
}

// UpdateIssueTicket 更新ticket
func (b *Bundle) UpdateIssueTicket(updateReq apistructs.IssueUpdateRequest, issueID uint64) error {
	host, err := b.urls.DOP()
	if err != nil {
		return err
	}
	hc := b.hc
	var buf bytes.Buffer
	resp, err := hc.Put(host).Path("/api/issues/"+strconv.FormatInt(int64(issueID), 10)).
		Header(httputil.InternalHeader, "bundle").JSONBody(&updateReq).Do().Body(&buf)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return apierrors.ErrInvoke.InternalError(
			fmt.Errorf("failed to update Ticket, status code: %d, body: %v",
				resp.StatusCode(),
				buf.String(),
			))
	}

	return nil
}

// UpdateIssueTicketUser 更新ticket，带User-ID
func (b *Bundle) UpdateIssueTicketUser(UserID string, updateReq apistructs.IssueUpdateRequest, issueID uint64) error {
	host, err := b.urls.DOP()
	if err != nil {
		return err
	}
	hc := b.hc
	var buf bytes.Buffer
	resp, err := hc.Put(host).Path("/api/issues/"+strconv.FormatInt(int64(issueID), 10)).
		Header(httputil.UserHeader, UserID).JSONBody(&updateReq).Do().Body(&buf)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return apierrors.ErrInvoke.InternalError(
			fmt.Errorf("failed to update Ticket, status code: %d, body: %v",
				resp.StatusCode(),
				buf.String(),
			))
	}

	return nil
}

func (b *Bundle) GetIssueStateBelong(req apistructs.IssueStateRelationGetRequest) ([]apistructs.IssueStateState, error) {
	host, err := b.urls.DOP()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var isp apistructs.IssueStateTypeBelongResponse
	resp, err := hc.Get(host).Path("/api/issues/actions/get-state-belong").
		Header(httputil.InternalHeader, "bundle").
		Param("projectID", strconv.FormatUint(req.ProjectID, 10)).
		Param("issueType", string(req.IssueType)).
		Do().JSON(&isp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !isp.Success {
		return nil, toAPIError(resp.StatusCode(), isp.Error)
	}
	return isp.Data, nil
}

func (b *Bundle) GetIssueStatesByID(req []int64) ([]apistructs.IssueStatus, error) {
	host, err := b.urls.DOP()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var isp apistructs.IssueStateNameGetResponse
	resp, err := hc.Get(host).Path("/api/issues/actions/get-state-name").
		Header(httputil.InternalHeader, "bundle").JSONBody(&req).
		Do().JSON(&isp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !isp.Success {
		return nil, toAPIError(resp.StatusCode(), isp.Error)
	}

	return isp.Data, nil
}

// UpdateIssuePanelIssue 更新事件所属看板
func (b *Bundle) UpdateIssuePanelIssue(userID string, panelID, issueID, projectID int64) error {
	host, err := b.urls.DOP()
	if err != nil {
		return err
	}
	hc := b.hc

	var panel apistructs.IssuePanelIssuesCreateResponse
	resp, err := hc.Put(host).Path("/api/issues/actions/update-panel-issue").
		Header(httputil.InternalHeader, "bundle").
		Header("User-ID", userID).
		Param("panelID", strconv.FormatInt(panelID, 10)).
		Param("issueID", strconv.FormatInt(issueID, 10)).
		Param("projectID", strconv.FormatInt(projectID, 10)).
		Do().JSON(&panel)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !panel.Success {
		return toAPIError(resp.StatusCode(), panel.Error)
	}

	return nil
}
func (b *Bundle) GetIssuePanel(req apistructs.IssuePanelRequest) ([]apistructs.IssuePanelIssues, error) {
	host, err := b.urls.DOP()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var isp apistructs.IssuePanelGetResponse
	resp, err := hc.Get(host).Path("/api/issues/actions/get-panel").
		Param("projectID", strconv.FormatUint(req.ProjectID, 10)).
		Header("User-ID", req.UserID).
		Header(httputil.InternalHeader, "bundle").
		Do().JSON(&isp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !isp.Success {
		return nil, toAPIError(resp.StatusCode(), isp.Error)
	}

	return isp.Data, nil
}

func (b *Bundle) GetIssuePanelIssue(req apistructs.IssuePanelRequest) (*apistructs.IssuePanelIssueIDs, error) {
	host, err := b.urls.DOP()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var isp apistructs.IssuePanelIssuesGetResponse
	resp, err := hc.Get(host).Path("/api/issues/actions/get-panel-issue").
		Param("panelID", strconv.FormatInt(req.PanelID, 10)).
		Params(req.UrlQueryString()).
		Header(httputil.InternalHeader, "bundle").
		Header("User-ID", req.UserID).
		Do().JSON(&isp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !isp.Success {
		return nil, toAPIError(resp.StatusCode(), isp.Error)
	}

	return isp.Data, nil
}

func (b *Bundle) CreateIssuePanel(req apistructs.IssuePanelRequest) (int64, error) {
	host, err := b.urls.DOP()
	if err != nil {
		return 0, err
	}
	hc := b.hc

	var isp apistructs.IssuePanelIssuesCreateResponse
	resp, err := hc.Post(host).Path("/api/issues/actions/create-panel").
		Header("User-ID", req.UserID).
		Header(httputil.InternalHeader, "bundle").JSONBody(&req).
		Do().JSON(&isp)
	if err != nil {
		return 0, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !isp.Success {
		return 0, toAPIError(resp.StatusCode(), isp.Error)
	}

	return isp.Data, nil
}

func (b *Bundle) DeleteIssuePanel(req apistructs.IssuePanelRequest) (*apistructs.IssuePanel, error) {
	host, err := b.urls.DOP()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var isp apistructs.IssuePanelDeleteResponse
	resp, err := hc.Delete(host).Path("/api/issues/actions/delete-panel").
		Header(httputil.InternalHeader, "bundle").
		Header("User-ID", req.UserID).
		Param("panelID", strconv.FormatInt(req.PanelID, 10)).
		Param("projectID", strconv.FormatUint(req.ProjectID, 10)).
		Do().JSON(&isp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !isp.Success {
		return nil, toAPIError(resp.StatusCode(), isp.Error)
	}

	return isp.Data, nil
}

func (b *Bundle) UpdateIssuePanel(req apistructs.IssuePanelRequest) (int64, error) {
	host, err := b.urls.DOP()
	if err != nil {
		return 0, err
	}
	hc := b.hc

	var isp apistructs.IssuePanelIssuesCreateResponse
	resp, err := hc.Put(host).Path("/api/issues/actions/update-panel").
		Header(httputil.InternalHeader, "bundle").
		Header("User-ID", req.UserID).
		Param("panelID", strconv.FormatInt(req.PanelID, 10)).
		Param("PanelName", req.PanelName).
		Param("projectID", strconv.FormatUint(req.ProjectID, 10)).
		Do().JSON(&isp)
	if err != nil {
		return 0, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !isp.Success {
		return 0, toAPIError(resp.StatusCode(), isp.Error)
	}

	return isp.Data, nil
}
func (b *Bundle) GetIssueStage(orgID int64, issueType apistructs.IssueType) ([]apistructs.IssueStage, error) {
	host, err := b.urls.DOP()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var isp apistructs.IssueStageResponse
	resp, err := hc.Get(host).Path("/api/issues/action/get-stage").
		Header(httputil.InternalHeader, "bundle").
		Param("orgID", strconv.FormatInt(orgID, 10)).
		Param("issueType", string(issueType)).
		Do().JSON(&isp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !isp.Success {
		return nil, toAPIError(resp.StatusCode(), isp.Error)
	}

	return isp.Data, nil
}
