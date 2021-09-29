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

package member

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/erda-project/erda-proto-go/msp/member/pb"
	projectpb "github.com/erda-project/erda-proto-go/msp/tenant/project/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/common/errors"
)

type memberService struct {
	p *provider
}

func (m memberService) ListMemberRoles(ctx context.Context, request *pb.ListMemberRolesRequest) (*pb.ListMemberRolesResponse, error) {
	var projectId string
	instance, err := m.p.mspTenantDB.QueryTenant(request.ScopeId)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if instance == nil {
		tenant, err := m.p.instanceDB.GetInstanceByTenantGroup(request.ScopeId)
		if err != nil {
			return nil, errors.NewInternalServerError(err)
		}
		option := make(map[string]string)
		if tenant != nil {
			err = json.Unmarshal([]byte(tenant.Options), &option)
			if err != nil {
				return nil, errors.NewInternalServerError(err)
			}
			projectId = option["projectId"]
		}
	} else {
		projectId = instance.RelatedProjectId
	}
	project, err := m.p.ProjectServer.GetProject(ctx, &projectpb.GetProjectRequest{
		ProjectID: projectId,
	})
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	orgIdStr := apis.GetOrgID(ctx)
	orgId, err := strconv.Atoi(orgIdStr)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if project.Data.Type == "DOP" {
		scopeId, err := strconv.Atoi(projectId)
		if err != nil {
			return nil, errors.NewInternalServerError(err)
		}
		roleList, err := m.p.bdl.ListMemberRoles(apistructs.ListScopeManagersByScopeIDRequest{
			ScopeType: apistructs.ScopeType(request.ScopeType),
			ScopeID:   int64(scopeId),
		}, int64(orgId))
		if err != nil {
			return nil, errors.NewInternalServerError(err)
		}
		roleListResult := &pb.ListMemberRolesResponse{
			Data: &pb.RoleList{
				List:  make([]*pb.RoleInfo, 0),
				Total: int64(roleList.Total),
			},
		}
		for _, v := range roleList.List {
			if v.Role == "" && v.Name == "" {
				roleListResult.Data.Total--
				continue
			}
			role := &pb.RoleInfo{
				Role: v.Role,
				Name: v.Name,
			}
			roleListResult.Data.List = append(roleListResult.Data.List, role)
		}
		return roleListResult, nil
	}
	return &pb.ListMemberRolesResponse{
		Data: &pb.RoleList{
			List: []*pb.RoleInfo{
				{
					Role: "Owner",
					Name: "项目所有者",
				},
				{
					Role: "Lead",
					Name: "研发主管",
				},
				{
					Role: "Dev",
					Name: "开发工程师",
				},
			},
			Total: 3,
		},
	}, nil
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
	userId := apis.GetUserID(ctx)
	deleteReq.IdentityInfo.UserID = userId
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
