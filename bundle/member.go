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
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/http/httputil"
)

// GetMemberByToken get member by token
func (b *Bundle) GetMemberByToken(request *apistructs.GetMemberByTokenRequest) (*apistructs.Member, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var fetchResp apistructs.GetMemberByTokenResponse
	resp, err := hc.Get(host).Path("/api/members/actions/get-by-token").
		Param("token", request.Token).
		Header(httputil.InternalHeader, "bundle").Do().JSON(&fetchResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !fetchResp.Success {
		return nil, toAPIError(resp.StatusCode(), fetchResp.Error)
	}

	return &fetchResp.Data, nil
}

func (b *Bundle) ListMembers(req apistructs.MemberListRequest) ([]apistructs.Member, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var r apistructs.MemberListResponse
	request := hc.Get(host).Path("/api/members").
		Header(httputil.InternalHeader, "bundle").
		Param("scopeType", string(req.ScopeType)).
		Param("scopeId", strconv.FormatInt(req.ScopeID, 10)).
		Param("q", req.Q).
		Param("pageNo", strconv.Itoa(req.PageNo)).
		Param("pageSize", strconv.Itoa(req.PageSize))
	for _, role := range req.Roles {
		request = request.Param("roles", role)
	}
	for _, label := range req.Labels {
		request = request.Param("label", label)
	}
	resp, err := request.Do().JSON(&r)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !r.Success {
		return nil, toAPIError(resp.StatusCode(), r.Error)
	}
	return r.Data.List, nil
}

// UpdateMemberUserInfo 更新成员的用户信息
func (b *Bundle) UpdateMemberUserInfo(req apistructs.MemberUserInfoUpdateRequest) error {
	host, err := b.urls.CMDB()
	if err != nil {
		return err
	}
	hc := b.hc

	var buf bytes.Buffer
	resp, err := hc.Put(host).Path("/api/members/actions/update-userinfo").
		Header(httputil.InternalHeader, "bundle").JSONBody(&req).Do().Body(&buf)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return apierrors.ErrInvoke.InternalError(
			fmt.Errorf("failed to update member's user info, status code: %d, body: %v",
				resp.StatusCode(),
				buf.String(),
			))
	}
	return nil
}

// DeleteMember 移除成员
func (b *Bundle) DeleteMember(req apistructs.MemberRemoveRequest) error {
	host, err := b.urls.CMDB()
	if err != nil {
		return err
	}
	hc := b.hc

	var buf bytes.Buffer
	resp, err := hc.Post(host).Path("/api/members/actions/remove").Header("User-ID", req.UserID).
		Header(httputil.InternalHeader, "bundle").JSONBody(&req).Do().Body(&buf)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return apierrors.ErrInvoke.InternalError(
			fmt.Errorf("failed to remove member, status code: %d, body: %v",
				resp.StatusCode(),
				buf.String(),
			))
	}
	return nil
}

// DestroyUsers 删除用户一切成员信息
func (b *Bundle) DestroyUsers(req apistructs.MemberDestroyRequest) error {
	host, err := b.urls.CMDB()
	if err != nil {
		return err
	}
	hc := b.hc

	var buf bytes.Buffer
	resp, err := hc.Post(host).Path("/api/members/actions/destroy").
		Header(httputil.InternalHeader, "bundle").JSONBody(&req).Do().Body(&buf)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return apierrors.ErrInvoke.InternalError(
			fmt.Errorf("failed to destroy member, status code: %d, body: %v",
				resp.StatusCode(),
				buf.String(),
			))
	}
	return nil
}

// ListMemberRolesByUser 查看某个用户在一个scope下的权限
func (b *Bundle) ListMemberRolesByUser(req apistructs.ListMemberRolesByUserRequest) (
	*apistructs.UserRoleListResponseData, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var memberResp apistructs.ListMemberRolesByUserResponse
	resp, err := hc.Get(host).Path("/api/members/actions/list-user-roles").
		Header(httputil.InternalHeader, "bundle").
		Header("lang", "en-US").
		Param("userId", req.UserID).
		Param("scopeType", string(req.ScopeType)).
		Param("parentId", strconv.FormatInt(req.ParentID, 10)).
		Param("pageNo", strconv.Itoa(req.PageNo)).
		Param("pageSize", strconv.Itoa(req.PageSize)).
		Do().JSON(&memberResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return nil, apierrors.ErrInvoke.InternalError(
			fmt.Errorf("failed to list user's role, status code: %d, body: %v",
				resp.StatusCode(),
				memberResp.Error,
			))
	}

	return &memberResp.Data, nil
}

// GetAllOrganizational 获取所有的组织架构
func (b *Bundle) GetAllOrganizational() (*apistructs.GetAllOrganizationalData, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := httpclient.New(
		httpclient.WithTimeout(time.Second, time.Minute*3),
	)

	var organizationalResp apistructs.GetAllOrganizationalResponse
	resp, err := hc.Get(host).Path("/api/members/actions/get-all-organizational").
		Header(httputil.InternalHeader, "bundle").Do().JSON(&organizationalResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return nil, apierrors.ErrInvoke.InternalError(
			fmt.Errorf("failed to list all organizational, status code: %d, body: %v",
				resp.StatusCode(),
				organizationalResp.Error,
			))
	}

	return &organizationalResp.Data, nil
}
