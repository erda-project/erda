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
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmdb/conf"
	"github.com/erda-project/erda/modules/cmdb/model"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/httputil"
	"github.com/erda-project/erda/pkg/loop"
	"github.com/erda-project/erda/pkg/strutil"
)

var (
	// job/deployment列表的任务存在时间，默认7天
	TaskCleanDurationTimestamp int64 = 7 * 24 * 60 * 60
)

// CreateOrg 创建企业
func (e *Endpoints) CreateOrg(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrCreateOrg.NotLogin().ToResp(), nil
	}

	// 操作鉴权, 只有系统管理员可创建企业
	if !e.member.IsAdmin(userID.String()) {
		return apierrors.ErrCreateOrg.AccessDenied().ToResp(), nil
	}

	// 检查body合法性
	if r.Body == nil {
		return apierrors.ErrCreateOrg.MissingParameter("body").ToResp(), nil
	}

	var orgCreateReq apistructs.OrgCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&orgCreateReq); err != nil {
		return apierrors.ErrCreateOrg.InvalidParameter(err).ToResp(), nil
	}
	logrus.Infof("request body: %+v", orgCreateReq)

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

	org, err := e.org.CreateWithEvent(userID.String(), orgCreateReq)
	if err != nil {
		return apierrors.ErrCreateOrg.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(e.convertToOrgDTO(*org))
}

// 更新企业Ingress
func (e *Endpoints) UpdateOrgIngress(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 检查orgID合法性
	orgID, err := strutil.Atoi64(vars["orgID"])
	if err != nil {
		return apierrors.ErrUpdateOrgIngress.InvalidParameter(err).ToResp(), nil
	}

	org, err := e.org.Get(orgID)
	if err != nil {
		return apierrors.ErrUpdateOrgIngress.InvalidParameter(err).ToResp(), nil
	}
	if err := e.bdl.CreateOrUpdateComponentIngress(apistructs.ComponentIngressUpdateRequest{
		ComponentName: "ui",
		ComponentPort: 80,
		IngressName:   org.Name + "-org",
		Routes: []apistructs.IngressRoute{
			{
				Domain: org.Name + "-org." + conf.RootDomain(),
				Path:   "/",
			},
		},
		RouteOptions: apistructs.RouteOptions{
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/proxy-body-size":   "0",
				"nginx.ingress.kubernetes.io/enable-access-log": "false",
			},
		},
	}); err != nil {
		return apierrors.ErrCreateOrg.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(e.convertToOrgDTO(*org))
}

// CreateOrgPublisher 创建发布商
func (e *Endpoints) CreateOrgPublisher(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrUpdateOrg.NotLogin().ToResp(), nil
	}

	orgID, err := strconv.ParseUint(vars["orgID"], 10, 64)
	if err != nil {
		return apierrors.ErrUpdateOrg.InvalidParameter(err).ToResp(), nil
	}

	// check permission
	if access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.OrgScope,
		ScopeID:  orgID,
		Resource: apistructs.OrgResource,
		Action:   apistructs.CreateAction,
	}); err != nil || !access {
		if err != nil {
			logrus.Errorf("failed to check permission when create org publisher, err: %v", err)
		}
		return apierrors.ErrUpdateOrg.AccessDenied().ToResp(), nil
	}

	org, err := e.org.Get(int64(orgID))
	if err != nil {
		return apierrors.ErrUpdateOrg.InvalidParameter(err).ToResp(), nil
	}
	orgDto := e.convertToOrgDTO(*org)
	if orgDto.PublisherID != 0 {
		return apierrors.ErrUpdateOrg.InvalidParameter(errors.New("org already has publisher")).ToResp(), nil
	}

	publisherName := ""
	if r.Method == "POST" {
		var publisherCreateReq apistructs.CreateOrgPublisherRequest
		if err := json.NewDecoder(r.Body).Decode(&publisherCreateReq); err != nil {
			return apierrors.ErrCreateOrg.InvalidParameter(err).ToResp(), nil
		}
		publisherName = publisherCreateReq.Name
	} else {
		publisherName = r.URL.Query().Get("name")
	}

	if publisherName == "" {
		return apierrors.ErrUpdateOrg.InvalidParameter(errors.New("name is empty")).ToResp(), nil
	}

	pub := &apistructs.PublisherCreateRequest{
		Name:          publisherName,
		PublisherType: "ORG",
		OrgID:         orgDto.ID,
	}

	_, err = e.publisher.Create(userID.String(), pub)
	if err != nil {
		return apierrors.ErrUpdateOrg.InvalidParameter(err).ToResp(), nil
	}

	return httpserver.OkResp("")
}

// UpdateOrg 更新企业
func (e *Endpoints) UpdateOrg(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrUpdateOrg.NotLogin().ToResp(), nil
	}

	// 检查orgID合法性
	orgID, err := strutil.Atoi64(vars["orgID"])
	if err != nil {
		return apierrors.ErrUpdateOrg.InvalidParameter(err).ToResp(), nil
	}

	// 检查body合法性
	if r.Body == nil {
		return apierrors.ErrUpdateOrg.MissingParameter("body").ToResp(), nil
	}
	var orgUpdateReq apistructs.OrgUpdateRequestBody
	if err := json.NewDecoder(r.Body).Decode(&orgUpdateReq); err != nil {
		return apierrors.ErrUpdateOrg.InvalidParameter(err).ToResp(), nil
	}
	logrus.Infof("request body: %+v", orgUpdateReq)

	// 操作鉴权
	req := apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.OrgScope,
		ScopeID:  uint64(orgID),
		Resource: apistructs.OrgResource,
		Action:   apistructs.UpdateAction,
	}
	if access, err := e.permission.CheckPermission(&req); err != nil || !access {
		return apierrors.ErrUpdateOrg.AccessDenied().ToResp(), nil
	}
	// 更新企业信息至DB
	org, err := e.org.UpdateWithEvent(orgID, orgUpdateReq)
	if err != nil {
		return apierrors.ErrUpdateOrg.InternalError(err).ToResp(), nil
	}
	orgDTO := e.convertToOrgDTO(*org)
	// 传递了publisherName并且当前org没有publisher才创建
	if orgUpdateReq.PublisherName != "" && orgDTO.PublisherID == 0 {
		publisherID, err := e.publisher.Create(userID.String(), &apistructs.PublisherCreateRequest{
			Name:          orgUpdateReq.PublisherName,
			PublisherType: "ORG",
			OrgID:         uint64(org.ID),
		})
		if err != nil {
			return apierrors.ErrUpdateOrg.InternalError(err).ToResp(), err
		}
		orgDTO.PublisherID = publisherID
	}

	return httpserver.OkResp(orgDTO)
}

// GetOrg 获取企业详情
func (e *Endpoints) GetOrg(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		userID user.ID
		org    *model.Org
		err    error
	)
	// 检查orgID合法性
	orgStr := vars["idOrName"]
	orgID, _ := strutil.Atoi64(orgStr)
	if orgID == 0 { // 按 orgName 查询
		org, err = e.org.GetByName(orgStr)
		if err != nil {
			return apierrors.ErrGetOrg.InternalError(err).ToResp(), nil
		}
	} else { // 按 orgID 查询
		org, err = e.org.Get(orgID)
		if err != nil {
			return apierrors.ErrGetOrg.InternalError(err).ToResp(), nil
		}
	}

	internalClient := r.Header.Get(httputil.InternalHeader)
	if internalClient == "" {
		userID, err = user.GetUserID(r)
		if err != nil {
			return apierrors.ErrGetOrg.NotLogin().ToResp(), nil
		}
		// 操作鉴权
		req := apistructs.PermissionCheckRequest{
			UserID:   userID.String(),
			Scope:    apistructs.OrgScope,
			ScopeID:  uint64(org.ID),
			Resource: apistructs.OrgResource,
			Action:   apistructs.GetAction,
		}
		if access, err := e.permission.CheckPermission(&req); err != nil || !access {
			return apierrors.ErrGetOrg.AccessDenied().ToResp(), nil
		}
	}

	result := e.convertToOrgDTO(*org)
	// 不是内部调用不返回配置信息
	if internalClient == "" {
		result.HidePassword()
	}
	// 封装org返回结构
	return httpserver.OkResp(result)
}

// DeleteOrg 删除企业
func (e *Endpoints) DeleteOrg(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDeleteOrg.NotLogin().ToResp(), nil
	}

	var org *model.Org
	// 检查orgID合法性
	orgStr := vars["idOrName"]
	orgID, _ := strutil.Atoi64(orgStr)
	if orgID == 0 { // 按 orgName 查询
		org, err = e.org.GetByName(orgStr)
		if err != nil {
			return apierrors.ErrDeleteOrg.InternalError(err).ToResp(), nil
		}
	} else { // 按 orgID 查询
		org, err = e.org.Get(orgID)
		if err != nil {
			return apierrors.ErrDeleteOrg.InternalError(err).ToResp(), nil
		}
	}

	if !identityInfo.IsInternalClient() {
		// 操作鉴权
		req := apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.OrgScope,
			ScopeID:  uint64(org.ID),
			Resource: apistructs.OrgResource,
			Action:   apistructs.DeleteAction,
		}
		if access, err := e.permission.CheckPermission(&req); err != nil || !access {
			return apierrors.ErrDeleteOrg.AccessDenied().ToResp(), nil
		}
	}

	// 删除企业
	if err = e.org.Delete(org.ID); err != nil {
		return errorresp.ErrResp(err)
	}

	// 封装org返回结构
	return httpserver.OkResp(e.convertToOrgDTO(*org))
}

// ListOrg 查询企业 GET /api/orgs?key=xxx(按企业名称过滤)
func (e *Endpoints) ListOrg(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	all, orgIDs, err := e.getOrgPermissions(r)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	// 获取请求参数
	req, err := getOrgListParam(r)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	var currentOrgID int64
	v := r.Header.Get(httputil.OrgHeader)
	if v != "" {
		// ignore convert error
		currentOrgID, err = strutil.Atoi64(v)
		if err != nil {
			return apierrors.ErrListOrg.InvalidParameter(strutil.Concat("orgID: ", v)).ToResp(), nil
		}
	}
	if currentOrgID == 0 {
		// compatible TODO: refactor it
		userID, _ := user.GetUserID(r)
		currentOrgID, _ = e.org.GetCurrentOrgByUser(userID.String())
	}
	var (
		total int
		orgs  []model.Org
	)
	// TODO: move this logic to service layer
	if all {
		total, orgs, err = e.org.SearchByName(req.Q, req.PageNo, req.PageSize)
	} else {
		total, orgs, err = e.org.ListByIDsAndName(orgIDs, req.Q, req.PageNo, req.PageSize)
	}
	if err != nil {
		logrus.Warnf("failed to get orgs, (%v)", err)
		return apierrors.ErrListOrg.InternalError(err).ToResp(), nil
	}

	// 封装成API所需格式
	orgDTOs := make([]apistructs.OrgDTO, 0, len(orgs))
	var selected bool // 是否已选中某企业flag
	for _, org := range orgs {
		orgDTO := e.convertToOrgDTO(org)
		if (currentOrgID == 0 || currentOrgID == org.ID) && !selected {
			orgDTO.Selected = true
			selected = true
		}
		orgDTO.HidePassword()
		orgDTOs = append(orgDTOs, orgDTO)
	}
	if !selected && len(orgDTOs) > 0 { // 用户选中某个企业后被踢出此企业后
		orgDTOs[0].Selected = true
	}

	return httpserver.OkResp(apistructs.PagingOrgDTO{
		List:  orgDTOs,
		Total: total,
	})
}

// Get public orgs
func (e *Endpoints) ListPublicOrg(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	_, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrListPublicOrg.NotLogin().ToResp(), nil
	}

	req, err := getOrgListParam(r)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	total, orgs, err := e.org.SearchPublicOrgsByName(req.Q, req.PageNo, req.PageSize)
	if err != nil {
		return apierrors.ErrListPublicOrg.InternalError(err).ToResp(), nil
	}

	orgDTOs := make([]apistructs.OrgDTO, 0, len(orgs))
	for _, org := range orgs {
		orgDTO := e.convertToOrgDTO(org)
		orgDTO.HidePassword()
		orgDTOs = append(orgDTOs, orgDTO)
	}

	return httpserver.OkResp(apistructs.PagingOrgDTO{
		List:  orgDTOs,
		Total: total,
	})
}

// GetOrgByDomain 通过域名查询企业
func (e *Endpoints) GetOrgByDomain(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		return apierrors.ErrGetOrg.MissingParameter("domain").ToResp(), nil
	}

	org, err := e.org.GetOrgByDomain(domain)
	if err != nil {
		return apierrors.ErrGetOrg.InternalError(err).ToResp(), nil
	}
	if org == nil {
		return httpserver.OkResp(nil)
	}
	return httpserver.OkResp(e.convertToOrgDTO(*org, domain))
}

// Deprecated: no need to store selection
// ChangeCurrentOrg 切换当前企业
func (e *Endpoints) ChangeCurrentOrg(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrChangeOrg.InvalidParameter(err).ToResp(), nil
	}

	if r.Body == nil {
		return apierrors.ErrChangeOrg.MissingParameter("body is nil").ToResp(), nil
	}
	var req apistructs.OrgChangeRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrChangeOrg.InvalidParameter(err).ToResp(), nil
	}

	if err = e.org.ChangeCurrentOrg(userID.String(), &req); err != nil {
		return apierrors.ErrChangeOrg.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(true)
}

// CreateOrgClusterRelation 创建企业集群关联关系
func (e *Endpoints) CreateOrgClusterRelation(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrRelateCluster.NotLogin().ToResp(), nil
	}

	// 解析body参数
	if r.Body == nil {
		return apierrors.ErrRelateCluster.MissingParameter("body").ToResp(), nil
	}
	var req apistructs.OrgClusterRelationCreateRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrRelateCluster.InvalidParameter(err).ToResp(), nil
	}

	if req.OrgID == 0 {
		return apierrors.ErrRelateCluster.MissingParameter("org id").ToResp(), nil
	}
	if req.ClusterName == "" {
		return apierrors.ErrRelateCluster.MissingParameter("cluster name").ToResp(), nil
	}

	if err = e.org.RelateCluster(userID.String(), &req); err != nil {
		return apierrors.ErrRelateCluster.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp("success")
}

// ListAllOrgClusterRelation 获取所有企业对应集群关系
func (e *Endpoints) ListAllOrgClusterRelation(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	_, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrRelateCluster.NotLogin().ToResp(), nil
	}

	rels, err := e.org.ListAllOrgClusterRelation()
	if err != nil {
		return apierrors.ErrGetOrgClusterRelation.InternalError(err).ToResp(), nil
	}

	var relDTOs []apistructs.OrgClusterRelationDTO
	for _, rel := range rels {
		relDTOs = append(relDTOs, convertToOrgClusterRelationDTO(rel))
	}
	return httpserver.OkResp(relDTOs)
}

func (e *Endpoints) convertToOrgDTO(org model.Org, domains ...string) apistructs.OrgDTO {
	domain := ""
	if len(domains) > 0 {
		domain = domains[0]
	}
	domain_and_port := strutil.Split(domain, ":", true)
	port := ""
	if len(domain_and_port) > 1 {
		port = domain_and_port[1]
	}
	concat_domain := strutil.Concat(strutil.ToLower(org.Name), "-org.", conf.RootDomain())
	if port != "" {
		concat_domain = strutil.Concat(strutil.ToLower(org.Name), "-org.", conf.RootDomain(), ":", port)
	}

	orgDto := apistructs.OrgDTO{
		ID:          uint64(org.ID),
		Name:        org.Name,
		Desc:        org.Desc,
		Logo:        org.Logo,
		Locale:      org.Locale,
		Domain:      concat_domain,
		Creator:     org.UserID,
		OpenFdp:     org.OpenFdp,
		DisplayName: org.DisplayName,
		PublisherID: e.org.GetPublisherID(org.ID),
		Config: &apistructs.OrgConfig{
			EnableMS:                   org.Config.EnableMS,
			SMTPHost:                   org.Config.SMTPHost,
			SMTPUser:                   org.Config.SMTPUser,
			SMTPPassword:               org.Config.SMTPPassword,
			SMTPPort:                   org.Config.SMTPPort,
			SMTPIsSSL:                  org.Config.SMTPIsSSL,
			SMSKeyID:                   org.Config.SMSKeyID,
			SMSKeySecret:               org.Config.SMSKeySecret,
			SMSSignName:                org.Config.SMSSignName,
			SMSMonitorTemplateCode:     org.Config.SMSMonitorTemplateCode, // 监控单独的短信模版
			VMSKeyID:                   org.Config.VMSKeyID,
			VMSKeySecret:               org.Config.VMSKeySecret,
			VMSMonitorTtsCode:          org.Config.VMSMonitorTtsCode, // 监控单独的语音模版
			VMSMonitorCalledShowNumber: org.Config.VMSMonitorCalledShowNumber,
		},
		IsPublic: org.IsPublic,
		BlockoutConfig: apistructs.BlockoutConfig{
			BlockDEV:   org.BlockoutConfig.BlockDEV,
			BlockTEST:  org.BlockoutConfig.BlockTEST,
			BlockStage: org.BlockoutConfig.BlockStage,
			BlockProd:  org.BlockoutConfig.BlockProd,
		},
		EnableReleaseCrossCluster: org.Config.EnableReleaseCrossCluster,
		CreatedAt:                 org.CreatedAt,
		UpdatedAt:                 org.UpdatedAt,
	}
	if orgDto.DisplayName == "" {
		orgDto.DisplayName = orgDto.Name
	}
	return orgDto
}

func convertToOrgClusterRelationDTO(rel model.OrgClusterRelation) apistructs.OrgClusterRelationDTO {
	return apistructs.OrgClusterRelationDTO{
		ID:          uint64(rel.ID),
		OrgID:       rel.OrgID,
		OrgName:     rel.OrgName,
		ClusterID:   rel.ClusterID,
		ClusterName: rel.ClusterName,
		Creator:     rel.Creator,
		CreatedAt:   rel.CreatedAt,
		UpdatedAt:   rel.UpdatedAt,
	}
}

func getOrgListParam(r *http.Request) (*apistructs.OrgSearchRequest, error) {
	q := r.URL.Query().Get("q")
	if q == "" {
		// Deprecated: TODO: remove it
		q = r.URL.Query().Get("key")
	}
	// Get paging info
	pageSizeStr := r.URL.Query().Get("pageSize")
	if pageSizeStr == "" {
		pageSizeStr = "20"
	}
	pageNumStr := r.URL.Query().Get("pageNo")
	if pageNumStr == "" {
		pageNumStr = "1"
	}
	size, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return nil, apierrors.ErrListOrg.InvalidParameter(strutil.Concat("PageSize: ", pageSizeStr))
	}
	num, err := strconv.Atoi(pageNumStr)
	if err != nil {
		return nil, apierrors.ErrListOrg.InvalidParameter(strutil.Concat("PageNo: ", pageNumStr))
	}
	if num <= 0 {
		num = 1
	}
	if size <= 0 {
		size = 20
	}
	return &apistructs.OrgSearchRequest{
		Q:        q,
		PageNo:   num,
		PageSize: size,
	}, nil
}

func (e *Endpoints) getOrgPermissions(r *http.Request) (bool, []int64, error) {
	// first check user
	userID, _ := user.GetUserID(r)

	//verify calls from bundle
	if userID == "" && r.Header.Get(httputil.InternalHeader) != "" {
		return true, nil, nil
	}

	// 操作鉴权, 系统管理员可查询企业
	if e.member.IsAdmin(userID.String()) { // 系统管理员可查看所有企业列表
		return true, nil, nil
	} else { // 非系统管理员只能查看有权限的企业列表
		members, err := e.member.ListByScopeTypeAndUser(apistructs.OrgScope, userID.String())
		if err != nil {
			return false, nil, apierrors.ErrListOrg.InternalError(err)
		}
		var orgIDs []int64
		for i := range members {
			orgIDs = append(orgIDs, members[i].ScopeID)
		}
		return false, orgIDs, nil
	}
}

// TODO: need refactor
func (e *Endpoints) getPermOrgs(r *http.Request) (bool, []int64, error) {
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return false, nil, apierrors.ErrListOrg.NotLogin()
	}
	// 操作鉴权, 系统管理员可查询企业
	if e.member.IsAdmin(userID.String()) { // 系统管理员可查看所有企业列表
		return true, nil, nil
	} else { // 非系统管理员只能查看有权限的企业列表
		members, err := e.member.ListByScopeTypeAndUser(apistructs.OrgScope, userID.String())
		if err != nil {
			return false, nil, apierrors.ErrListOrg.InternalError(err)
		}
		var orgIDs []int64
		for i := range members {
			orgIDs = append(orgIDs, members[i].ScopeID)
		}
		return false, orgIDs, nil
	}
}

// ListOrgRunningTasks  指定企业获取服务或者job列表
func (e *Endpoints) ListOrgRunningTasks(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	org, err := e.getOrgByRequest(r)
	if err != nil {
		return apierrors.ErrListClusterAbnormalHosts.InvalidParameter("org id header").ToResp(), nil
	}

	reqParam, err := e.getRunningTasksListParam(r)
	if err != nil {
		return apierrors.ErrListClusterAbnormalHosts.InvalidParameter(err).ToResp(), nil
	}

	total, tasksResults, err := e.org.ListOrgRunningTasks(reqParam, org.ID)
	if err != nil {
		return apierrors.ErrListOrgRunningTasks.InternalError(err).ToResp(), nil
	}

	// insert userID
	userIDs := make([]string, 0, len(tasksResults))
	for _, task := range tasksResults {
		userIDs = append(userIDs, task.UserID)
	}

	return httpserver.OkResp(apistructs.OrgRunningTasksData{Total: total, Tasks: tasksResults},
		strutil.DedupSlice(userIDs, true))
}

func (e *Endpoints) getRunningTasksListParam(r *http.Request) (*apistructs.OrgRunningTasksListRequest, error) {
	// 获取type参数
	taskType := r.URL.Query().Get("type")
	if taskType == "" {
		return nil, errors.Errorf("type")
	}

	if taskType != "job" && taskType != "deployment" {
		return nil, errors.Errorf("type")
	}

	cluster := r.URL.Query().Get("cluster")
	projectName := r.URL.Query().Get("projectName")
	appName := r.URL.Query().Get("appName")
	status := r.URL.Query().Get("status")
	userID := r.URL.Query().Get("userID")
	env := r.URL.Query().Get("env")

	var (
		startTime int64
		endTime   int64
		pipeline  uint64
		err       error
	)
	pipelineID := r.URL.Query().Get("pipelineID")
	if pipelineID != "" {
		pipeline, err = strconv.ParseUint(pipelineID, 10, 64)
		if err != nil {
			return nil, errors.Errorf("convert pipelineID, (%+v)", err)
		}
	}

	// 获取时间范围
	startTimeStr := r.URL.Query().Get("startTime")
	if startTimeStr != "" {
		startTime, err = strutil.Atoi64(startTimeStr)
		if err != nil {
			return nil, err
		}
	}

	endTimeStr := r.URL.Query().Get("endTime")
	if endTimeStr != "" {
		endTime, err = strutil.Atoi64(endTimeStr)
		if err != nil {
			return nil, err
		}
	}

	// 获取pageNo参数
	pageNoStr := r.URL.Query().Get("pageNo")
	if pageNoStr == "" {
		pageNoStr = "1"
	}
	pageNo, err := strconv.Atoi(pageNoStr)
	if err != nil {
		return nil, errors.Errorf("invalid param, pageNo: %v", pageNoStr)
	}

	// 获取pageSize参数
	pageSizeStr := r.URL.Query().Get("pageSize")
	if pageSizeStr == "" {
		pageSizeStr = "20"
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return nil, errors.Errorf("invalid param, pageSize: %v", pageSizeStr)
	}

	return &apistructs.OrgRunningTasksListRequest{
		Cluster:     cluster,
		ProjectName: projectName,
		AppName:     appName,
		PipelineID:  pipeline,
		Status:      status,
		UserID:      userID,
		Env:         env,
		StartTime:   startTime,
		EndTime:     endTime,
		PageNo:      pageNo,
		PageSize:    pageSize,
		Type:        taskType,
	}, nil
}

// DealTaskEvents 接收任务的事件
func (e *Endpoints) DealTaskEvent(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		req           apistructs.PipelineTaskEvent
		runningTaskID int64
		err           error
	)
	if r.Body == nil {
		return apierrors.ErrDealTaskEvents.MissingParameter("body").ToResp(), nil
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrDealTaskEvents.InvalidParameter(err).ToResp(), nil
	}
	logrus.Debugf("ReceiveTaskEvents: request body: %+v", req)

	if req.Event == "pipeline_task" {
		if runningTaskID, err = e.org.DealReceiveTaskEvent(&req); err != nil {
			return apierrors.ErrDealTaskEvents.InvalidParameter(err).ToResp(), nil
		}
	} else if req.Event == "pipeline_task_runtime" {
		if runningTaskID, err = e.org.DealReceiveTaskRuntimeEvent(&req); err != nil {
			return apierrors.ErrDealTaskEvents.InvalidParameter(err).ToResp(), nil
		}
	}

	return httpserver.OkResp(runningTaskID)
}

// SyncTaskStatus 定时同步主机实际使用资源
func (e *Endpoints) SyncTaskStatus(interval time.Duration) {
	l := loop.New(loop.WithInterval(interval))
	l.Do(func() (bool, error) {
		// deal job resource
		jobs := e.db.ListRunningJobs()

		for _, job := range jobs {
			// 根据pipelineID获取task列表信息
			bdl := bundle.New(bundle.WithPipeline())
			pipelineInfo, err := bdl.GetPipeline(job.PipelineID)
			if err != nil {
				logrus.Errorf("failed to get pipeline info by pipelineID, pipelineID:%d, (%+v)", job.PipelineID, err)
				continue
			}

			for _, stage := range pipelineInfo.PipelineStages {
				for _, task := range stage.PipelineTasks {
					if task.ID == job.TaskID {
						if string(task.Status) != job.Status {
							job.Status = string(task.Status)

							// 更新数据库状态
							e.db.UpdateJobStatus(&job)
							logrus.Debugf("update job status, jobID:%d, status:%s", job.ID, job.Status)
						}
					}
				}
			}

		}

		// deal deployment resource
		deployments := e.db.ListRunningDeployments()

		for _, deployment := range deployments {
			// 根据pipelineID获取task列表信息
			bdl := bundle.New(bundle.WithPipeline())
			pipelineInfo, err := bdl.GetPipeline(deployment.PipelineID)
			if err != nil {
				logrus.Errorf("failed to get pipeline info by pipelineID, pipelineID:%d, (%+v)", deployment.PipelineID, err)
				continue
			}

			for _, stage := range pipelineInfo.PipelineStages {
				for _, task := range stage.PipelineTasks {
					if task.ID == deployment.TaskID {
						if string(task.Status) != deployment.Status {
							deployment.Status = string(task.Status)

							// 更新数据库状态
							e.db.UpdateDeploymentStatus(&deployment)
						}
					}
				}
			}

		}

		return false, nil
	})
}

// TaskClean 定时清理任务(job/deployment)资源
func (e *Endpoints) TaskClean(interval time.Duration) {
	l := loop.New(loop.WithInterval(interval))
	l.Do(func() (bool, error) {
		timeUnix := time.Now().Unix()
		fmt.Println(timeUnix)

		startTimestamp := timeUnix - TaskCleanDurationTimestamp

		startTime := time.Unix(startTimestamp, 0).Format("2006-01-02 15:04:05")

		// clean job resource
		jobs := e.db.ListExpiredJobs(startTime)

		for _, job := range jobs {
			err := e.db.DeleteJob(strconv.FormatUint(job.OrgID, 10), job.TaskID)
			if err != nil {
				err = e.db.DeleteJob(strconv.FormatUint(job.OrgID, 10), job.TaskID)
				if err != nil {
					logrus.Errorf("failed to delete job, job: %+v, (%+v)", job, err)
				}
			}
			logrus.Debugf("[clean] expired job: %+v", job)
		}

		// clean deployment resource
		deployments := e.db.ListExpiredDeployments(startTime)

		for _, deployment := range deployments {
			err := e.db.DeleteDeployment(strconv.FormatUint(deployment.OrgID, 10), deployment.TaskID)
			if err != nil {
				err = e.db.DeleteDeployment(strconv.FormatUint(deployment.OrgID, 10), deployment.TaskID)
				if err != nil {
					logrus.Errorf("failed to delete deployment, deployment: %+v, (%+v)", deployment, err)
				}
			}

			logrus.Debugf("[clean] expired deployment: %+v", deployment)
		}

		return false, nil
	})
}

// FetchOrgResources 获取企业资源情况
func (e *Endpoints) FetchOrgResources(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrFetchOrgResources.NotLogin().ToResp(), nil
	}
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return apierrors.ErrFetchOrgResources.NotLogin().ToResp(), nil
	}
	orgID, err := strconv.ParseUint(orgIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrFetchOrgResources.InvalidParameter("org id header").ToResp(), nil
	}

	if !identityInfo.IsInternalClient() {
		// 操作鉴权
		req := apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.OrgScope,
			ScopeID:  orgID,
			Resource: apistructs.ResourceInfoResource,
			Action:   apistructs.GetAction,
		}
		if access, err := e.permission.CheckPermission(&req); err != nil || !access {
			return apierrors.ErrFetchOrgResources.AccessDenied().ToResp(), nil
		}
	}

	orgResourceInfo, err := e.org.FetchOrgResources(orgID)
	if err != nil {
		return apierrors.ErrFetchOrgResources.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(orgResourceInfo)
}

func (e *Endpoints) SetReleaseCrossCluster(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrSetReleaseCrossCluster.NotLogin().ToResp(), nil
	}
	if !identityInfo.IsInternalClient() {
		req := apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.SysScope,
			Resource: apistructs.OrgResource,
			Action:   apistructs.UpdateAction,
		}
		if access, err := e.permission.CheckPermission(&req); err != nil || !access {
			return apierrors.ErrSetReleaseCrossCluster.AccessDenied().ToResp(), nil
		}
	}
	enable, err := strconv.ParseBool(r.URL.Query().Get("enable"))
	if err != nil {
		return apierrors.ErrSetReleaseCrossCluster.InvalidParameter("invalid bool query: enable").ToResp(), nil
	}
	orgID, err := strconv.ParseUint(vars["orgID"], 10, 64)
	if err != nil {
		return apierrors.ErrSetReleaseCrossCluster.InvalidParameter("orgID").ToResp(), nil
	}
	if err := e.org.SetReleaseCrossCluster(orgID, enable); err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(nil)
}

// GenVerifiCode 生成邀请成员加入企业的验证码
func (e *Endpoints) GenVerifiCode(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetOrgVerifiCode.NotLogin().ToResp(), nil
	}
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return apierrors.ErrGetOrgVerifiCode.MissingParameter("org-id is empty").ToResp(), nil
	}
	orgID, err := strconv.ParseUint(orgIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrGetOrgVerifiCode.InvalidParameter("org id header").ToResp(), nil
	}
	if !identityInfo.IsInternalClient() {
		req := apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.OrgScope,
			ScopeID:  orgID,
			Resource: apistructs.MemberResource,
			Action:   apistructs.CreateAction,
		}
		if access, err := e.permission.CheckPermission(&req); err != nil || !access {
			return apierrors.ErrGetOrgVerifiCode.AccessDenied().ToResp(), nil
		}
	}

	// 获取验证码
	verifiCode, err := e.org.GenVerifiCode(identityInfo, orgID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(map[string]string{"verifyCode": verifiCode})
}

// SetNotifyConfig 设置通知配置
func (e *Endpoints) SetNotifyConfig(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		userID user.ID
		err    error
	)
	// 检查orgID合法性
	orgID, err := strutil.Atoi64(vars["orgID"])
	if err != nil {
		return apierrors.ErrGetNotifyConfig.InvalidParameter(err).ToResp(), nil
	}

	// 检查body合法性
	if r.Body == nil {
		return apierrors.ErrUpdateOrg.MissingParameter("body").ToResp(), nil
	}
	var nCfgUpdateReq apistructs.NotifyConfigUpdateRequestBody
	if err := json.NewDecoder(r.Body).Decode(&nCfgUpdateReq); err != nil {
		return apierrors.ErrUpdateOrg.InvalidParameter(err).ToResp(), nil
	}

	internalClient := r.Header.Get(httputil.InternalHeader)
	if internalClient == "" {
		userID, err = user.GetUserID(r)
		if err != nil {
			return apierrors.ErrGetNotifyConfig.NotLogin().ToResp(), nil
		}
		// 操作鉴权
		req := apistructs.PermissionCheckRequest{
			UserID:   userID.String(),
			Scope:    apistructs.OrgScope,
			ScopeID:  uint64(orgID),
			Resource: apistructs.NotifyConfigResource,
			Action:   apistructs.UpdateAction,
		}
		if access, err := e.permission.CheckPermission(&req); err != nil || !access {
			return apierrors.ErrGetNotifyConfig.AccessDenied().ToResp(), nil
		}
	}

	if err := e.org.SetNotifyConfig(orgID, nCfgUpdateReq); err != nil {
		return apierrors.ErrGetNotifyConfig.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp("succes")
}

// GetNotifyConfig 获取通知配置
func (e *Endpoints) GetNotifyConfig(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// var (
	// userID user.ID
	// 	err error
	// )
	// 检查orgID合法性
	orgID, err := strutil.Atoi64(vars["orgID"])
	if err != nil {
		return apierrors.ErrGetNotifyConfig.InvalidParameter(err).ToResp(), nil
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

	result, err := e.org.GetNotifyConfig(orgID)
	if err != nil {
		return apierrors.ErrGetNotifyConfig.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(apistructs.NotifyConfigUpdateRequestBody{
		Config: result,
	})
}
