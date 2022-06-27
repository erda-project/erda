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

package org

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/legacy/services/apierrors"
	"github.com/erda-project/erda/internal/core/legacy/services/member"
	"github.com/erda-project/erda/internal/core/legacy/services/permission"
	"github.com/erda-project/erda/internal/core/legacy/utils"
	"github.com/erda-project/erda/internal/core/org/db"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/erda-project/erda/pkg/ucauth"
)

type Interface interface {
	pb.OrgServiceServer
	WithUc(uc *ucauth.UCClient) *provider
	WithMember(member *member.Member) *provider
	WithPermission(permission *permission.Permission) *provider
	ListOrgs(ctx context.Context, orgIDs []int64, req *pb.ListOrgRequest, all bool) (int, []*pb.Org, error)
}

// CreateOrg 创建企业
func (p *provider) CreateOrg(ctx context.Context, req *pb.CreateOrgRequest) (*pb.CreateOrgResponse, error) {
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, apierrors.ErrCreateOrg.NotLogin()
	}

	// 操作鉴权, 只有系统管理员可创建企业
	isAdmin := p.member.IsAdmin(userID)
	// when env: create_org_enabled is true, allow all people create org
	if !p.Cfg.CreateOrgEnabled {
		if !isAdmin {
			return nil, apierrors.ErrCreateOrg.AccessDenied()
		}
	}

	// check the org name is invalid
	if !utils.IsValidOrgName(req.Name) {
		return nil, apierrors.ErrCreateOrg.InvalidParameter(errors.Errorf("org name is invalid %s", req.Name))
	}

	// check if it is free org, currently only admin can create paid organizations
	if req.Type != apistructs.FreeOrgType.String() && !isAdmin {
		return nil, apierrors.ErrCreateOrg.AccessDenied()
	}
	// compatible logic, delete this after perfecting the logic of organization creation
	if req.Type == "" && isAdmin {
		req.Type = apistructs.EnterpriseOrgType.String()
	}

	logrus.Infof("request body: %+v", req)

	// 第一次创建企业的时候还没有集群，没有集群创建ingress会出错，先注释掉了
	// if err := e.bdl.CreateOrUpdateComponentIngress(apistructs.ComponentIngressUpdateRequest{
	// 	ComponentName: "ui",
	// 	ComponentPort: 80,
	// 	IngressName:   orgCreateReq.Name + "-org",
	// 	Routes: []apistructs.IngressRoute{
	// 		{
	// 			Domain: orgCreateReq.Name + "-org." + conf.RootDomain(),
	// 			Path:   "/",
	// 		},
	// 	},
	// 	RouteOptions: apistructs.RouteOptions{
	// 		Annotations: map[string]string{
	// 			"nginx.ingress.kubernetes.io/proxy-body-size":   "0",
	// 			"nginx.ingress.kubernetes.io/enable-access-log": "false",
	// 		},
	// 	},
	// }); err != nil {
	// 	return apierrors.ErrCreateOrg.InternalError(err).ToResp(), nil
	// }

	org, err := p.CreateWithEvent(req)
	if err != nil {
		return nil, apierrors.ErrCreateOrg.InternalError(err)
	}

	return &pb.CreateOrgResponse{Data: p.convertToOrgDTO(org)}, nil
}

// UpdateOrg 更新企业
func (p *provider) UpdateOrg(ctx context.Context, req *pb.UpdateOrgRequest) (*pb.UpdateOrgResponse, error) {
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, apierrors.ErrUpdateOrg.NotLogin()
	}

	orgID, err := strutil.Atoi64(req.OrgID)
	if err != nil {
		return nil, apierrors.ErrUpdateOrg.InvalidParameter(err)
	}

	logrus.Infof("request body: %+v", req)

	// 操作鉴权
	permReq := apistructs.PermissionCheckRequest{
		UserID:   userID,
		Scope:    apistructs.OrgScope,
		ScopeID:  uint64(orgID),
		Resource: apistructs.OrgResource,
		Action:   apistructs.UpdateAction,
	}
	if access, err := p.permission.CheckPermission(&permReq); err != nil || !access {
		return nil, apierrors.ErrUpdateOrg.AccessDenied()
	}
	// 更新企业信息至DB
	org, auditMessage, err := p.UpdateWithEvent(orgID, req)
	if err != nil {
		return nil, apierrors.ErrUpdateOrg.InternalError(err)
	}
	orgDTO := p.convertToOrgDTO(org)
	orgDTO.AuditMessage = auditMessage

	return &pb.UpdateOrgResponse{Data: orgDTO}, nil
}

// GetOrg 获取企业详情
func (p *provider) GetOrg(ctx context.Context, req *pb.GetOrgRequest) (*pb.GetOrgResponse, error) {
	var (
		userID string
		org    *db.Org
		err    error
	)
	// 检查orgID合法性
	orgStr := req.IdOrName
	orgID, _ := strutil.Atoi64(orgStr)
	if orgID == 0 { // 按 orgName 查询
		org, err = p.GetByName(orgStr)
		if err != nil {
			return nil, apierrors.ErrGetOrg.InternalError(err)
		}
	} else { // 按 orgID 查询
		org, err = p.Get(orgID)
		if err != nil {
			return nil, apierrors.ErrGetOrg.InternalError(err)
		}
	}

	internalClient := apis.GetInternalClient(ctx)
	if internalClient == "" {
		userID = apis.GetUserID(ctx)
		// 操作鉴权
		req := apistructs.PermissionCheckRequest{
			UserID:   userID,
			Scope:    apistructs.OrgScope,
			ScopeID:  uint64(org.ID),
			Resource: apistructs.OrgResource,
			Action:   apistructs.GetAction,
		}
		if access, err := p.permission.CheckPermission(&req); err != nil || !access {
			return nil, apierrors.ErrGetOrg.AccessDenied()
		}
	}

	result := p.convertToOrgDTO(org)
	// 不是内部调用不返回配置信息
	if internalClient == "" {
		HidePassword(result)
	}
	// 封装org返回结构
	return &pb.GetOrgResponse{Data: result}, nil
}

// DeleteOrg 删除企业
func (p *provider) DeleteOrg(ctx context.Context, req *pb.DeleteOrgRequest) (*pb.DeleteOrgResponse, error) {
	identityInfo := apis.GetIdentityInfo(ctx)
	if identityInfo == nil {
		return nil, apierrors.ErrDeleteOrg.NotLogin()
	}

	var (
		org *db.Org
		err error
	)
	// 检查orgID合法性
	orgStr := req.IdOrName
	orgID, _ := strutil.Atoi64(orgStr)
	if orgID == 0 { // 按 orgName 查询
		org, err = p.GetByName(orgStr)
		if err != nil {
			return nil, apierrors.ErrDeleteOrg.InternalError(err)
		}
	} else { // 按 orgID 查询
		org, err = p.Get(orgID)
		if err != nil {
			return nil, apierrors.ErrDeleteOrg.InternalError(err)
		}
	}

	if identityInfo.InternalClient == "" {
		// 操作鉴权
		req := apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.OrgScope,
			ScopeID:  uint64(org.ID),
			Resource: apistructs.OrgResource,
			Action:   apistructs.DeleteAction,
		}
		if access, err := p.permission.CheckPermission(&req); err != nil || !access {
			return nil, apierrors.ErrDeleteOrg.AccessDenied()
		}
	}

	// 删除企业
	if err = p.Delete(org.ID); err != nil {
		return nil, apierrors.ErrDeleteOrg.InternalError(err)
	}

	// 封装org返回结构
	return &pb.DeleteOrgResponse{Data: p.convertToOrgDTO(org)}, nil
}

// ListOrg 查询企业 GET /api/orgs?key=xxx(按企业名称过滤)
func (p *provider) ListOrg(ctx context.Context, req *pb.ListOrgRequest) (*pb.ListOrgResponse, error) {
	all, orgIDs, err := p.getOrgPermissions(ctx, req)
	if err != nil {
		return nil, apierrors.ErrListOrg.InternalError(err)
	}
	setOrgListParam(req)

	var (
		total int
		orgs  []*pb.Org
	)
	total, orgs, err = p.ListOrgs(ctx, orgIDs, req, all)
	if err != nil {
		logrus.Warnf("failed to get orgs, (%v)", err)
		return nil, apierrors.ErrListOrg.InternalError(err)
	}

	return &pb.ListOrgResponse{List: orgs, Total: int64(total)}, nil
}

func setOrgListParam(req *pb.ListOrgRequest) {
	if req.Q == "" {
		req.Q = req.Key
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}
	if req.PageNo == 0 {
		req.PageNo = 1
	}
	return
}

func (p *provider) getOrgPermissions(ctx context.Context, req *pb.ListOrgRequest) (bool, []int64, error) {
	internalClient := apis.GetInternalClient(ctx)
	userID := apis.GetUserID(ctx)
	// verify calls from bundle
	if userID == "" && internalClient != "" {
		return true, nil, nil
	}
	// 操作鉴权, 系统管理员可查询企业
	// Found that org is passed in the request header, even admin does not query all organizations
	if p.member.IsAdmin(userID) && (req.Org == "" || req.Org == "-") { // 系统管理员可查看所有企业列表
		return true, nil, nil
	} else { // 非系统管理员只能查看有权限的企业列表
		members, err := p.member.ListByScopeTypeAndUser(apistructs.OrgScope, userID)
		if err != nil {
			return false, nil, err
		}
		var orgIDs []int64
		for i := range members {
			orgIDs = append(orgIDs, members[i].ScopeID)
		}
		return false, orgIDs, nil
	}
}

// ListPublicOrg Get public orgs
func (p *provider) ListPublicOrg(ctx context.Context, req *pb.ListOrgRequest) (*pb.ListOrgResponse, error) {
	identityInfo := apis.GetIdentityInfo(ctx)
	if identityInfo == nil {
		return nil, apierrors.ErrListPublicOrg.NotLogin()
	}

	setOrgListParam(req)
	total, orgs, err := p.SearchPublicOrgsByName(req.Q, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return nil, apierrors.ErrListPublicOrg.InternalError(err)
	}

	orgDTOs := make([]*pb.Org, 0, len(orgs))
	for _, org := range orgs {
		orgDTO := p.convertToOrgDTO(&org)
		HidePassword(orgDTO)
		orgDTOs = append(orgDTOs, orgDTO)
	}

	return &pb.ListOrgResponse{List: orgDTOs, Total: int64(total)}, nil
}

// GetOrgByDomain 通过域名查询企业
func (p *provider) GetOrgByDomain(ctx context.Context, req *pb.GetOrgByDomainRequest) (*pb.GetOrgByDomainResponse, error) {
	if req.Domain == "" {
		return nil, apierrors.ErrGetOrg.MissingParameter("domain")
	}

	org, err := p.GetOrgByDomainAndOrgName(req.Domain, req.OrgName)
	if err != nil {
		return nil, apierrors.ErrGetOrg.InternalError(err)
	}

	return &pb.GetOrgByDomainResponse{Data: p.convertToOrgDTO(org)}, nil
}

// Deprecated: no need to store selection
// ChangeCurrentOrg 切换当前企业
func (p *provider) ChangeCurrentOrg(ctx context.Context, req *pb.ChangeCurrentOrgRequest) (*pb.ChangeCurrentOrgResponse, error) {
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, apierrors.ErrChangeOrg.NotLogin()
	}

	if err := p.changeCurrentOrg(userID, req); err != nil {
		return nil, apierrors.ErrChangeOrg.InternalError(err)
	}

	return &pb.ChangeCurrentOrgResponse{Data: true}, nil
}

// CreateOrgClusterRelation 创建企业集群关联关系
func (p *provider) CreateOrgClusterRelation(ctx context.Context, req *pb.OrgClusterRelationCreateRequest) (*pb.OrgClusterRelationCreateResponse, error) {
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, apierrors.ErrRelateCluster.NotLogin()
	}

	if req.OrgID == 0 {
		return nil, apierrors.ErrRelateCluster.MissingParameter("orgID")
	}
	if req.ClusterName == "" {
		return nil, apierrors.ErrRelateCluster.MissingParameter("clusterName")
	}

	if err := p.RelateCluster(userID, req); err != nil {
		return nil, apierrors.ErrRelateCluster.InternalError(err)
	}
	return &pb.OrgClusterRelationCreateResponse{Data: "success"}, nil
}

// ListOrgClusterRelation 获取所有企业对应集群关系
func (p *provider) ListOrgClusterRelation(ctx context.Context, req *pb.ListOrgClusterRelationRequest) (*pb.ListOrgClusterRelationResponse, error) {
	var (
		err  error
		rels []db.OrgClusterRelation
	)

	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, apierrors.ErrRelateCluster.NotLogin()
	}

	cluster := req.Cluster
	if cluster == "" {
		rels, err = p.ListAllOrgClusterRelation()
		if err != nil {
			return nil, apierrors.ErrGetOrgClusterRelation.InternalError(err)
		}
	} else {
		rels, err = p.dbClient.GetOrgClusterRelationsByCluster(cluster)
		if err != nil {
			return nil, apierrors.ErrGetOrgClusterRelationsByOrg.InvalidParameter(err)
		}
	}

	var relDTOs []*pb.OrgClusterRelation
	for _, rel := range rels {
		relDTOs = append(relDTOs, convertToOrgClusterRelationDTO(rel))
	}
	return &pb.ListOrgClusterRelationResponse{Data: relDTOs}, nil
}

func (p *provider) SetReleaseCrossCluster(ctx context.Context, req *pb.SetReleaseCrossClusterRequest) (*pb.SetReleaseCrossClusterResponse, error) {
	identityInfo := apis.GetIdentityInfo(ctx)
	if identityInfo == nil {
		return nil, apierrors.ErrSetReleaseCrossCluster.NotLogin()
	}
	if identityInfo.InternalClient == "" {
		req := apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.SysScope,
			Resource: apistructs.OrgResource,
			Action:   apistructs.UpdateAction,
		}
		if access, err := p.permission.CheckPermission(&req); err != nil || !access {
			return nil, apierrors.ErrSetReleaseCrossCluster.AccessDenied()
		}
	}

	orgID, err := strconv.ParseUint(req.OrgID, 10, 64)
	if err != nil {
		return nil, apierrors.ErrSetReleaseCrossCluster.InvalidParameter("orgID")
	}
	if err := p.setReleaseCrossCluster(orgID, req.Enable); err != nil {
		return nil, apierrors.ErrSetReleaseCrossCluster.InternalError(err)
	}
	return &pb.SetReleaseCrossClusterResponse{}, nil
}

// GenVerifyCode 生成邀请成员加入企业的验证码
func (p *provider) GenVerifyCode(ctx context.Context, req *pb.GenVerifyCodeRequest) (*pb.GenVerifyCodeResponse, error) {
	// 鉴权
	identityInfo := apis.GetIdentityInfo(ctx)
	if identityInfo == nil {
		return nil, apierrors.ErrGetOrgVerifiCode.NotLogin()
	}
	orgIDStr := apis.GetOrgID(ctx)
	if orgIDStr == "" {
		return nil, apierrors.ErrGetOrgVerifiCode.MissingParameter("orgID is empty")
	}
	orgID, err := strconv.ParseUint(orgIDStr, 10, 64)
	if err != nil {
		return nil, apierrors.ErrGetOrgVerifiCode.InvalidParameter("orgID header")
	}
	if identityInfo.InternalClient == "" {
		req := apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.OrgScope,
			ScopeID:  orgID,
			Resource: apistructs.MemberResource,
			Action:   apistructs.CreateAction,
		}
		if access, err := p.permission.CheckPermission(&req); err != nil || !access {
			return nil, apierrors.ErrGetOrgVerifiCode.AccessDenied()
		}
	}

	// 获取验证码
	verifyCode, err := p.genVerifyCode(identityInfo, orgID)
	if err != nil {
		return nil, apierrors.ErrGetOrgVerifiCode.InternalError(err)
	}
	return &pb.GenVerifyCodeResponse{Data: map[string]string{"verifyCode": verifyCode}}, nil
}

// SetNotifyConfig 设置通知配置
func (p *provider) SetNotifyConfig(ctx context.Context, req *pb.SetNotifyConfigRequest) (*pb.SetNotifyConfigResponse, error) {
	var err error
	// 检查orgID合法性
	orgID, err := strutil.Atoi64(req.OrgID)
	if err != nil {
		return nil, apierrors.ErrGetNotifyConfig.InvalidParameter(err)
	}

	identityInfo := apis.GetIdentityInfo(ctx)
	if identityInfo == nil {
		return nil, apierrors.ErrGetNotifyConfig.NotLogin()
	}

	if identityInfo.InternalClient == "" {
		if identityInfo.UserID == "" {
			return nil, apierrors.ErrGetNotifyConfig.NotLogin()
		}
		// 操作鉴权
		req := apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.OrgScope,
			ScopeID:  uint64(orgID),
			Resource: apistructs.NotifyConfigResource,
			Action:   apistructs.UpdateAction,
		}
		if access, err := p.permission.CheckPermission(&req); err != nil || !access {
			return nil, apierrors.ErrGetNotifyConfig.AccessDenied()
		}
	}

	if err := p.setNotifyConfig(orgID, req); err != nil {
		return nil, apierrors.ErrGetNotifyConfig.InternalError(err)
	}

	return &pb.SetNotifyConfigResponse{Data: "success"}, nil
}

// GetNotifyConfig 获取通知配置
func (p *provider) GetNotifyConfig(ctx context.Context, req *pb.GetNotifyConfigRequest) (*pb.GetNotifyConfigResponse, error) {
	// var (
	// userID user.ID
	// 	err error
	// )
	// 检查orgID合法性
	orgID, err := strutil.Atoi64(req.OrgID)
	if err != nil {
		return nil, apierrors.ErrGetNotifyConfig.InvalidParameter(err)
	}

	// internalClient := r.Header.Get(httputil.InternalHeader)
	// if internalClient == "" {
	// 	userID, err = user.GetUserID(r)
	// 	if err != nil {
	// 		return apierrors.ErrGetNotifyConfig.NotLogin().ToResp(), nil
	// 	}
	// 	// 操作鉴权
	// 	req := apistructs.PermissionCheckRequest{
	// 		UserID:   userID.String(),
	// 		Scope:    apistructs.OrgScope,
	// 		ScopeID:  uint64(orgID),
	// 		Resource: apistructs.NotifyConfigResource,
	// 		Action:   apistructs.GetAction,
	// 	}
	// 	if access, err := e.permission.CheckPermission(&req); err != nil || !access {
	// 		return apierrors.ErrGetNotifyConfig.AccessDenied().ToResp(), nil
	// 	}
	// }

	result, err := p.getNotifyConfig(orgID)
	if err != nil {
		return nil, apierrors.ErrGetNotifyConfig.InternalError(err)
	}

	return &pb.GetNotifyConfigResponse{Data: &pb.NotifyConfig{Config: result}}, nil
}

// GetOrgClusterRelationsByOrg get orgClusters relation by orgID
func (p *provider) GetOrgClusterRelationsByOrg(ctx context.Context, req *pb.GetOrgClusterRelationsByOrgRequest) (*pb.GetOrgClusterRelationsByOrgResponse, error) {
	identity := apis.GetIdentityInfo(ctx)
	if identity == nil {
		return nil, apierrors.ErrGetOrgClusterRelationsByOrg.NotLogin()
	}

	orgID, err := strutil.Atoi64(req.OrgID)
	if err != nil {
		return nil, apierrors.ErrGetOrgClusterRelationsByOrg.InvalidParameter(err)
	}

	orgList, err := p.dbClient.GetOrgClusterRelationsByOrg(orgID)
	if err != nil {
		return nil, apierrors.ErrGetOrgClusterRelationsByOrg.InvalidParameter(err)
	}

	orgDtoList := make([]*pb.OrgClusterRelation, 0, len(orgList))
	for _, v := range orgList {
		orgDtoList = append(orgDtoList, convertToOrgClusterRelationDTO(v))
	}
	return &pb.GetOrgClusterRelationsByOrgResponse{Data: orgDtoList}, nil
}

// DereferenceCluster 解除关联集群关系
func (p *provider) DereferenceCluster(ctx context.Context, req *pb.DereferenceClusterRequest) (*pb.DereferenceClusterResponse, error) {
	identity := apis.GetIdentityInfo(ctx)
	if identity == nil {
		return nil, apierrors.ErrDereferenceCluster.NotLogin()
	}
	orgIDStr := req.OrgID
	if orgIDStr == "" {
		return nil, apierrors.ErrDereferenceCluster.MissingParameter("orgID")
	}
	var (
		orgID int64
		err   error
	)
	if orgID, err = strutil.Atoi64(orgIDStr); err != nil {
		return nil, apierrors.ErrListCluster.InvalidParameter(err)
	}
	clusterName := req.ClusterName
	if clusterName == "" {
		return nil, apierrors.ErrDereferenceCluster.MissingParameter("clusterName")
	}
	permReq := apistructs.DereferenceClusterRequest{
		OrgID:   orgID,
		Cluster: clusterName,
	}
	if err := p.member.CheckPermission(identity.UserID, apistructs.OrgScope, permReq.OrgID); err != nil {
		return nil, apierrors.ErrDereferenceCluster.InternalError(err)
	}
	if err := p.dereferenceCluster(identity.UserID, &permReq); err != nil {
		return nil, apierrors.ErrDereferenceCluster.InternalError(err)
	}

	return &pb.DereferenceClusterResponse{Data: "success"}, nil
}

func convertToOrgClusterRelationDTO(rel db.OrgClusterRelation) *pb.OrgClusterRelation {
	return &pb.OrgClusterRelation{
		ID:          uint64(rel.ID),
		OrgId:       rel.OrgID,
		OrgName:     rel.OrgName,
		ClusterID:   rel.ClusterID,
		ClusterName: rel.ClusterName,
		Creator:     rel.Creator,
		CreatedAt:   timestamppb.New(rel.CreatedAt),
		UpdatedAt:   timestamppb.New(rel.UpdatedAt),
	}
}
