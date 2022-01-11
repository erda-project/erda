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

package notifygroup

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-proto-go/msp/apm/notifygroup/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/common/errors"
)

type notifyGroupService struct {
	p *provider
}

func (n *notifyGroupService) GetProjectIdByScopeId(scopeId string) (string, error) {
	projectId := ""
	var err error
	projectId, err = n.p.monitorDB.SelectProjectIdByTk(scopeId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			instance, err := n.p.mspTenantDB.QueryTenant(scopeId)
			if err != nil {
				return "", err
			}
			if instance != nil {
				projectId = instance.RelatedProjectId
			}
		} else {
			return projectId, err
		}
	}
	return projectId, nil
}

func (n *notifyGroupService) CreateNotifyGroup(ctx context.Context, request *pb.CreateNotifyGroupRequest) (*pb.CreateNotifyGroupResponse, error) {
	projectId, err := n.GetProjectIdByScopeId(request.ScopeId)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if projectId == "" {
		return nil, errors.NewInternalServerError(fmt.Errorf("Query project record by scopeid is empty scopeId is %v", request.ScopeId))
	}
	userId := apis.GetUserID(ctx)
	orgId := apis.GetOrgID(ctx)
	label := map[string]string{
		"member_scopeID":   projectId,
		"member_scopeType": "project",
	}
	data, err := json.Marshal(label)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	request.Label = string(data)
	data, err = json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	createReq := &apistructs.CreateNotifyGroupRequest{}
	err = json.Unmarshal(data, createReq)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	resp, err := n.p.bdl.CreateNotifyGroup(orgId, userId, createReq)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.CreateNotifyGroupResponse{
		Data: &pb.NotifyGroup{},
	}
	data, err = json.Marshal(resp)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	err = json.Unmarshal(data, result.Data)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return result, nil
}

func (n *notifyGroupService) auditContextInfo(groupId int64, orgId string) (string, string, string, uint64, error) {
	notifyGroup, err := n.p.bdl.GetNotifyGroup(groupId, orgId)
	if err != nil {
		return "", "", "", 0, err
	}
	workspace, err := n.getWorkSpace(notifyGroup.Data.ScopeID)
	if err != nil {
		return "", "", "", 0, err
	}
	projectName, auditProjectId, err := n.getProjectInfo(notifyGroup.Data.ScopeID)
	if err != nil {
		return "", "", "", 0, err
	}
	return projectName, workspace, notifyGroup.Data.Name, auditProjectId, nil
}

func (n *notifyGroupService) getProjectInfo(scopeId string) (string, uint64, error) {
	projectIdStr, err := n.GetProjectIdByScopeId(scopeId)
	if err != nil {
		return "", 0, errors.NewInternalServerError(err)
	}
	if projectIdStr == "" {
		return "", 0, errors.NewInternalServerError(fmt.Errorf("Query project record by scopeid is empty scopeId is %v", scopeId))
	}
	projectId, err := strconv.Atoi(projectIdStr)
	if err != nil {
		return "", 0, errors.NewInternalServerError(err)
	}
	auditProjectId := uint64(projectId)
	project, err := n.p.bdl.GetProject(auditProjectId)
	if err != nil {
		{
			return "", 0, errors.NewInternalServerError(err)
		}
	}
	return project.Name, auditProjectId, nil
}

func (n *notifyGroupService) getWorkSpace(scopeId string) (string, error) {
	workspace, err := n.p.mspTenantDB.GetTenantWorkspaceByTenantID(scopeId)
	if err != nil {
		return "", errors.NewInternalServerError(err)
	}
	if workspace == "" {
		workspace, err = n.p.monitorDB.GetWorkspaceByTK(scopeId)
		if err != nil {
			return "", errors.NewInternalServerError(err)
		}
	}
	return workspace, nil
}

func (n *notifyGroupService) QueryNotifyGroup(ctx context.Context, request *pb.QueryNotifyGroupRequest) (*pb.QueryNotifyGroupResponse, error) {
	orgId := apis.GetOrgID(ctx)
	queryReq := &apistructs.QueryNotifyGroupRequest{
		PageNo:      request.PageNo,
		PageSize:    request.PageSize,
		ScopeType:   request.ScopeType,
		ScopeID:     request.ScopeId,
		Label:       request.Label,
		ClusterName: request.ClusterName,
		Names:       request.Names,
	}
	resp, err := n.p.bdl.QueryNotifyGroup(orgId, queryReq)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	data, err := json.Marshal(resp.Data.List)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.QueryNotifyGroupResponse{
		Data: &pb.QueryNotifyGroupData{
			List:  make([]*pb.NotifyGroup, 0),
			Total: int64(resp.Data.Total),
		},
		UserIDs: resp.UserIDs,
	}
	err = json.Unmarshal(data, &result.Data.List)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return result, nil
}

func (n *notifyGroupService) GetNotifyGroup(ctx context.Context, request *pb.GetNotifyGroupRequest) (*pb.GetNotifyGroupResponse, error) {
	orgId := apis.GetOrgID(ctx)
	resp, err := n.p.bdl.GetNotifyGroup(request.GroupID, orgId)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.GetNotifyGroupResponse{
		Data:    &pb.NotifyGroup{},
		UserIDs: resp.UserIDs,
	}
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	err = json.Unmarshal(data, result.Data)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	data, err = json.Marshal(resp.UserInfo)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return result, nil
}

func (n *notifyGroupService) UpdateNotifyGroup(ctx context.Context, request *pb.UpdateNotifyGroupRequest) (*pb.UpdateNotifyGroupResponse, error) {
	orgID := apis.GetOrgID(ctx)
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	updateReq := &apistructs.UpdateNotifyGroupRequest{}
	err = json.Unmarshal(data, updateReq)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	resp, err := n.p.bdl.UpdateNotifyGroup(request.GroupID, orgID, updateReq)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	data, err = json.Marshal(resp)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.UpdateNotifyGroupResponse{
		Data: &pb.NotifyGroup{},
	}
	err = json.Unmarshal(data, result.Data)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	projectName, workspace, notifyGroupName, auditProjectId, err := n.auditContextInfo(request.GroupID, orgID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	userId := apis.GetUserID(ctx)
	user, err := n.p.bdl.GetCurrentUser(userId)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	auditContext := map[string]interface{}{
		"projectName":     projectName,
		"workspace":       workspace,
		"notifyGroupName": notifyGroupName,
		"userName":        user.Name,
	}
	audit := apistructs.ToAudit(apistructs.ProjectScope, userId, apistructs.UpdateNotifyGroup, auditProjectId, auditContext)
	if err := n.p.bdl.CreateAuditEvent(&apistructs.AuditCreateRequest{Audit: audit}); err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return result, nil
}

func (n *notifyGroupService) GetNotifyGroupDetail(ctx context.Context, request *pb.GetNotifyGroupDetailRequest) (*pb.GetNotifyGroupDetailResponse, error) {
	orgID := apis.GetOrgID(ctx)
	userID := apis.GetUserID(ctx)
	orgId, err := strconv.Atoi(orgID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	resp, err := n.p.bdl.GetNotifyGroupDetail(request.GroupID, int64(orgId), userID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.GetNotifyGroupDetailResponse{
		Data: &pb.NotifyGroupDetail{},
	}
	err = json.Unmarshal(data, result.Data)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return result, nil
}

func (n *notifyGroupService) DeleteNotifyGroup(ctx context.Context, request *pb.DeleteNotifyGroupRequest) (*pb.DeleteNotifyGroupResponse, error) {
	orgID := apis.GetOrgID(ctx)
	resp, err := n.p.bdl.DeleteNotifyGroup(request.GroupID, orgID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.DeleteNotifyGroupResponse{
		Data: &pb.NotifyGroup{},
	}
	err = json.Unmarshal(data, result.Data)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	projectName, workspace, notifyGroupName, auditProjectId, err := n.auditContextInfo(request.GroupID, orgID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	userId := apis.GetUserID(ctx)
	user, err := n.p.bdl.GetCurrentUser(userId)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	auditContext := map[string]interface{}{
		"projectName":     projectName,
		"workspace":       workspace,
		"notifyGroupName": notifyGroupName,
		"userName":        user.Name,
	}
	audit := apistructs.ToAudit(apistructs.ProjectScope, userId, apistructs.DeleteNotifyGroup, auditProjectId, auditContext)
	if err := n.p.bdl.CreateAuditEvent(&apistructs.AuditCreateRequest{Audit: audit}); err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return result, nil
}
