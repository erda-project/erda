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

package member

import (
	"context"
	"encoding/json"
	"github.com/erda-project/erda-proto-go/msp/apm/member/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/common/errors"
	"strconv"
)

type memberService struct {
	p *provider
}

func (m memberService) ListMemberRoles(ctx context.Context, request *pb.ListMemberRolesRequest) (*pb.ListMemberRolesResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	roleReq := apistructs.ListScopeManagersByScopeIDRequest{}
	err = json.Unmarshal(data, &roleReq)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	orgId := apis.GetOrgID(ctx)
	orgID, err := strconv.ParseInt(orgId, 10, 64)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	roleList, err := m.p.bdl.ListMemberRoles(roleReq, orgID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.ListMemberRolesResponse{
		Data: &pb.RoleList{
			List: make([]*pb.RoleInfo, 0),
		},
	}
	data, err = json.Marshal(roleList.List)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	err = json.Unmarshal(data, &result.Data.List)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result.Data.Total = int64(roleList.Total)
	return result, nil
}

func (m memberService) DeleteMember(ctx context.Context, request *pb.DeleteMemberRequest) (*pb.DeleteMemberResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	deleteReq := apistructs.MemberRemoveRequest{}
	err = json.Unmarshal(data, &deleteReq)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	err = m.p.bdl.DeleteMember(deleteReq)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (m memberService) CreateOrUpdateMember(ctx context.Context, request *pb.CreateOrUpdateMemberRequest) (*pb.CreateOrUpdateMemberResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	memberReq := apistructs.MemberAddRequest{}
	err = json.Unmarshal(data, &memberReq)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	userId := apis.GetUserID(ctx)
	err = m.p.bdl.AddMember(memberReq, userId)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return nil, nil
}

func (m memberService) ListMember(ctx context.Context, request *pb.ListMemberRequest) (*pb.ListMemberResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	listMemberReq := apistructs.MemberListRequest{}
	err = json.Unmarshal(data, &listMemberReq)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	listTotal, err := m.p.bdl.ListMembersAndTotal(listMemberReq)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.ListMemberResponse{
		Data: &pb.MemberList{
			List: make([]*pb.Member, 0),
		},
	}
	data, err = json.Marshal(listTotal)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	err = json.Unmarshal(data, result.Data)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return result, nil
}

func (m memberService) ScopeRoleAccess(ctx context.Context, request *pb.ScopeRoleAccessRequest) (*pb.ScopeRoleAccessResponse, error) {
	listReq := apistructs.ScopeRoleAccessRequest{
		Scope: apistructs.Scope{
			ID:   request.Scope.Id,
			Type: apistructs.ScopeType(request.Scope.Type),
		},
	}
	userId := apis.GetUserID(ctx)
	list, err := m.p.bdl.ScopeRoleAccessList(userId, &listReq)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	data, err := json.Marshal(list)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := pb.ScopeRoleAccessResponse{
		Access:                   false,
		Roles:                    make([]string, 0),
		PermissionList:           make([]*pb.ScopeResource, 0),
		ResourceRoleList:         make([]*pb.ScopeResource, 0),
		Exist:                    false,
		ContactsWhenNoPermission: make([]string, 0),
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &result, nil
}
