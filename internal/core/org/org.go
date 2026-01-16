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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda-proto-go/core/org/pb"
	tokenpb "github.com/erda-project/erda-proto-go/core/token/pb"
	userpb "github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/legacy/conf"
	"github.com/erda-project/erda/internal/core/legacy/model"
	"github.com/erda-project/erda/internal/core/legacy/services/apierrors"
	"github.com/erda-project/erda/internal/core/legacy/types"
	"github.com/erda-project/erda/internal/core/org/db"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/filehelper"
	"github.com/erda-project/erda/pkg/numeral"
	"github.com/erda-project/erda/pkg/oauth2/tokenstore/mysqltokenstore"
	"github.com/erda-project/erda/pkg/strutil"
)

const SECRECT_PLACEHOLDER = "******"

// CreateWithEvent 创建企业 & 发送创建事件
func (p *provider) CreateWithEvent(createReq *pb.CreateOrgRequest) (*db.Org, error) {
	// 创建企业
	org, err := p.Create(createReq)
	if err != nil {
		return nil, err
	}

	ev := &apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			Event:         bundle.OrgEvent,
			Action:        bundle.CreateAction,
			OrgID:         strconv.FormatInt(org.ID, 10),
			ProjectID:     "-1",
			ApplicationID: "-1",
			TimeStamp:     time.Now().Format("2006-01-02 15:04:05"),
		},
		Sender:  bundle.SenderCoreServices,
		Content: *org,
	}
	// 发送企业创建事件
	if err = p.bdl.CreateEvent(ev); err != nil {
		logrus.Warnf("failed to send org create event, (%v)", err)
	}
	return org, nil
}

// Create 创建企业处理逻辑
func (p *provider) Create(createReq *pb.CreateOrgRequest) (*db.Org, error) {
	if len(createReq.Admins) == 0 {
		return nil, errors.Errorf("failed to create org(param admins is empty)")
	}
	if createReq.DisplayName == "" {
		createReq.DisplayName = createReq.Name
	}
	if conf.RedirectPathList[strutil.Concat("/", createReq.Name)] {
		return nil, errors.Errorf("Org name in legacy redirect paths")
	}
	// name 校验
	if err := strutil.Validate(createReq.Name,
		strutil.NoChineseValidator,
		strutil.MaxLenValidator(50),
		strutil.AlphaNumericDashUnderscoreValidator,
	); err != nil {
		return nil, err
	}

	// 校验企业名唯一
	existOrg, err := p.dbClient.GetOrgByName(createReq.Name)
	if err != nil && err != db.ErrNotFoundOrg {
		return nil, err
	}
	if existOrg != nil {
		return nil, errors.Errorf("org name already exist")
	}

	// 添加企业至DB
	userID := createReq.Admins[0]
	org := &db.Org{
		Name:        createReq.Name,
		DisplayName: createReq.DisplayName,
		Desc:        createReq.Desc,
		Logo:        createReq.Logo,
		Locale:      createReq.Locale,
		UserID:      userID,
		Type:        createReq.Type,
		Status:      "ACTIVE",
		IsPublic:    createReq.IsPublic,
	}
	if err := p.dbClient.CreateOrg(org); err != nil {
		logrus.Warnf("failed to insert info to db, (%v)", err)
		return nil, errors.Errorf("failed to insert org info to db")
	}

	// 新增企业权限记录
	resp, err := p.uc.GetUser(
		apis.WithInternalClientContext(context.Background(), discover.SvcCoreServices),
		&userpb.GetUserRequest{UserID: userID},
	)
	if err != nil {
		logrus.Warnf("failed to query user info, (%v)", err)
	} else {
		user := resp.Data
		_, err := p.TokenService.CreateToken(context.Background(), &tokenpb.CreateTokenRequest{
			Scope:     string(apistructs.OrgScope),
			ScopeId:   strconv.FormatInt(org.ID, 10),
			Type:      mysqltokenstore.PAT.String(),
			CreatorId: userID,
		})
		if err != nil {
			logrus.Warnf("failed to create token, (%v)", err)
		}

		member := model.Member{
			ScopeType:  apistructs.OrgScope,
			ScopeID:    org.ID,
			ScopeName:  org.Name,
			UserID:     userID,
			Email:      user.Email,
			Mobile:     user.Phone,
			Name:       user.Name,
			Nick:       user.Nick,
			Avatar:     user.AvatarURL,
			UserSyncAt: time.Now(),
			OrgID:      org.ID,
		}
		if err = p.dbClient.DBClient.CreateMember(&member); err != nil {
			logrus.Warnf("failed to insert member info to db, (%v)", err)
		}
		if err = p.dbClient.DBClient.CreateMemberExtra(&model.MemberExtra{
			UserID:        userID,
			ScopeType:     apistructs.OrgScope,
			ScopeID:       org.ID,
			ResourceKey:   apistructs.RoleResourceKey,
			ResourceValue: types.RoleOrgManager,
		}); err != nil {
			logrus.Warnf("failed to insert member extra to db, (%v)", err)
		}
	}

	return org, nil
}

// UpdateWithEvent 更新企业 & 发送更新事件
func (p *provider) UpdateWithEvent(orgID int64, updateReq *pb.UpdateOrgRequest) (*db.Org, *pb.AuditMessage, error) {
	if updateReq.DisplayName == "" {
		updateReq.DisplayName = updateReq.Name
	}
	org, auditMessage, err := p.Update(orgID, updateReq)
	if err != nil {
		return nil, nil, err
	}

	ev := &apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			Event:         bundle.OrgEvent,
			Action:        bundle.UpdateAction,
			OrgID:         strconv.FormatInt(org.ID, 10),
			ProjectID:     "-1",
			ApplicationID: "-1",
			TimeStamp:     time.Now().Format("2006-01-02 15:04:05"),
		},
		Sender:  bundle.SenderCoreServices,
		Content: *org,
	}
	// 发送企业更新事件
	if err = p.bdl.CreateEvent(ev); err != nil {
		logrus.Warnf("failed to send org update event, (%v)", err)
	}

	return org, auditMessage, nil
}

// Update 更新企业
func (p *provider) Update(orgID int64, updateReq *pb.UpdateOrgRequest) (*db.Org, *pb.AuditMessage, error) {
	// 检查待更新的org是否存在
	org, err := p.dbClient.GetOrg(orgID)
	if err != nil {
		logrus.Warnf("failed to find org when update org, (%v)", err)
		return nil, nil, errors.Errorf("failed to find org when update org")
	}
	auditMessage := getAuditMessage(org, updateReq)

	// 更新企业元信息，企业名称暂不可更改
	org.Desc = updateReq.Desc
	org.Logo = updateReq.Logo
	org.DisplayName = updateReq.DisplayName
	org.Locale = updateReq.Locale
	org.IsPublic = updateReq.IsPublic
	if updateReq.Config != nil {
		org.Config.SMTPHost = updateReq.Config.SmtpHost
		org.Config.SMTPPort = updateReq.Config.SmtpPort
		org.Config.SMTPUser = updateReq.Config.SmtpUser
		org.Config.SMTPIsSSL = updateReq.Config.SmtpIsSSL
		org.Config.SMSKeyID = updateReq.Config.SmsKeyID
		org.Config.SMSSignName = updateReq.Config.SmsSignName
		if updateReq.Config.SmsKeySecret != "" && updateReq.Config.SmsKeySecret != apistructs.SECRECT_PLACEHOLDER {
			org.Config.SMSKeySecret = updateReq.Config.SmsKeySecret
		}
		if updateReq.Config.SmtpPassword != "" && updateReq.Config.SmtpPassword != apistructs.SECRECT_PLACEHOLDER {
			org.Config.SMTPPassword = updateReq.Config.SmtpPassword
		}
	}
	if updateReq.BlockoutConfig != nil {
		org.BlockoutConfig.BlockDEV = updateReq.BlockoutConfig.BlockDev
		org.BlockoutConfig.BlockTEST = updateReq.BlockoutConfig.BlockTest
		org.BlockoutConfig.BlockStage = updateReq.BlockoutConfig.BlockStage
		org.BlockoutConfig.BlockProd = updateReq.BlockoutConfig.BlockProd
	}

	// 更新企业信息至DB
	if err = p.dbClient.UpdateOrg(&org); err != nil {
		logrus.Warnf("failed to update org, (%v)", err)
		return nil, nil, errors.Errorf("failed to update org")
	}
	return &org, auditMessage, nil
}

// Get 获取企业
func (p *provider) Get(orgID int64) (*db.Org, error) {
	// 检查org是否存在
	org, err := p.dbClient.GetOrg(orgID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch org")
	}
	if org.DisplayName == "" {
		org.DisplayName = org.Name
	}
	return &org, nil
}

// Delete 删除企业
func (p *provider) Delete(orgID int64) error {
	// 检查企业下是否有项目
	projects, err := p.dbClient.ListProjectByOrgID(uint64(orgID))
	if err != nil {
		return err
	}
	if len(projects) > 0 {
		return errors.Errorf("project is not empty")
	}

	relations, err := p.dbClient.GetOrgClusterRelationsByOrg(orgID)
	if err != nil {
		return err
	}
	if len(relations) > 0 {
		return errors.Errorf("cluster relation exist")
	}

	// 删除企业下的成员及权限
	if err = p.dbClient.DeleteMembersByScope(apistructs.OrgScope, orgID); err != nil {
		return errors.Errorf("failed to delete members, (%v)", err)
	}
	if err = p.dbClient.DeleteMemberExtraByScope(apistructs.OrgScope, orgID); err != nil {
		return errors.Errorf("failed to delete members extra, (%v)", err)
	}

	return p.dbClient.DeleteOrg(orgID)
}

// GetByName 获取企业
func (p *provider) GetByName(orgName string) (*db.Org, error) {
	// 检查org是否存在
	org, err := p.dbClient.GetOrgByName(orgName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch org")
	}
	return org, nil
}

// SearchByName 按企业名称过滤
func (p *provider) SearchByName(name string, pageNo, pageSize int) (int, []db.Org, error) {
	total, orgs, err := p.dbClient.GetOrgsByParam(name, pageNo, pageSize)
	if err != nil {
		logrus.Warnf("failed to get orgs, (%v)", err)
		return 0, nil, err
	}
	return total, orgs, nil
}

// SearchPublicOrgsByName Search public orgs
func (p *provider) SearchPublicOrgsByName(name string, pageNo, pageSize int) (int, []db.Org, error) {
	total, orgs, err := p.dbClient.GetPublicOrgsByParam(name, pageNo, pageSize)
	if err != nil {
		return 0, nil, err
	}
	return total, orgs, nil
}

// ListByIDsAndName 根据IDs列表 & name 获取企业列表
func (p *provider) ListByIDsAndName(orgIDs []int64, name string, pageNo, pageSize int) (int, []db.Org, error) {
	return p.dbClient.GetOrgsByIDsAndName(orgIDs, name, pageNo, pageSize)
}

func (p *provider) ListOrgs(ctx context.Context, orgIDs []int64, req *pb.ListOrgRequest, all bool) (int, []*pb.Org, error) {
	var (
		total int
		orgs  []db.Org
		err   error
	)
	if all {
		total, orgs, err = p.SearchByName(req.Q, int(req.PageNo), int(req.PageSize))
	} else {
		total, orgs, err = p.ListByIDsAndName(orgIDs, req.Q, int(req.PageNo), int(req.PageSize))
	}
	if err != nil {
		logrus.Warnf("failed to get orgs, (%v)", err)
		return 0, nil, err
	}

	dtos, err := p.coverOrgsToDto(ctx, orgs)
	if err != nil {
		return 0, nil, err
	}
	return total, dtos, nil
}

// ChangeCurrentOrg 切换用户当前所属企业
func (p *provider) changeCurrentOrg(userID string, req *pb.ChangeCurrentOrgRequest) error {
	// 检查用户是否匹配
	if req.UserID != userID {
		return errors.Errorf("user id doesn't match")
	}
	// 检查企业是否存在
	if _, err := p.dbClient.GetOrg(int64(req.OrgID)); err != nil {
		return err
	}

	orgID, err := p.dbClient.GetCurrentOrgByUser(req.UserID)
	if err != nil || orgID == 0 { // 若当前登录用户currentOrg记录不存在
		currentOrg := &model.CurrentOrg{
			UserID: req.UserID,
			OrgID:  int64(req.OrgID),
		}
		return p.dbClient.CreateCurrentOrg(currentOrg)
	}

	return p.dbClient.UpdateCurrentOrg(req.UserID, int64(req.OrgID))
}

// GetCurrentOrgByUser 根据userID获取用户当前关联企业
func (p *provider) GetCurrentOrgByUser(userID string) (int64, error) {
	return p.dbClient.GetCurrentOrgByUser(userID)
}

// List 获取所有企业列表
func (p *provider) List() ([]db.Org, error) {
	return p.dbClient.GetOrgList()
}

// GetOrgByDomainAndOrgName Org by domain and org name
func (p *provider) GetOrgByDomainAndOrgName(domain, orgName string) (*db.Org, error) {
	if orgName == "" {
		return p.GetOrgByDomainFromDB(domain)
	}
	org, err := p.dbClient.GetOrgByName(orgName)
	if err != nil {
		if err != db.ErrNotFoundOrg {
			return nil, err
		}
		// Not found, search by domain
		org, err = p.GetOrgByDomainFromDB(domain)
		if err != nil {
			if err != db.ErrNotFoundOrg {
				return nil, err
			}
			return nil, nil
		}
	}
	return org, nil
}

// GetOrgByDomainFromDB 通过域名获取企业
func (p *provider) GetOrgByDomainFromDB(domain string) (*db.Org, error) {
	if domain != "" && conf.OrgWhiteList[domain] {
		return nil, nil
	}
	for _, rootDomain := range conf.RootDomainList() {
		if orgName := orgNameRetriever(domain, rootDomain); orgName != "" {
			return p.dbClient.GetOrgByName(orgName)
		}
	}
	return nil, apierrors.ErrGetOrg.NotFound()
}

// Search org name in domain
func orgNameRetriever(domain, rootDomain string) string {
	suf := strutil.Concat(".", rootDomain)
	domain_and_port := strutil.Split(domain, ":", true)
	domain = domain_and_port[0]
	if strutil.HasSuffixes(domain, suf) {
		orgName := strutil.TrimSuffixes(domain, suf)
		if strutil.HasSuffixes(orgName, "-org") {
			orgName = strutil.TrimSuffixes(orgName, "-org")
		}
		return orgName
	}
	return ""
}

// RelateCluster 关联集群，创建企业集群关联关系
func (p *provider) RelateCluster(userID string, req *pb.OrgClusterRelationCreateRequest) error {
	org, err := p.dbClient.GetOrg(int64(req.OrgID))
	if err != nil {
		return err
	}
	if org.ID == 0 {
		return errors.Errorf("org not found")
	}
	if req.OrgName == "" {
		req.OrgName = org.Name
	} else if org.Name != req.OrgName {
		return errors.Errorf("org info doesn't match")
	}
	cluster, err := p.bdl.GetCluster(req.ClusterName)
	if err != nil {
		return err
	}
	if cluster == nil {
		return errors.Errorf("cluster not found")
	}
	// 若企业集群关系已存在，则返回
	relation, err := p.dbClient.GetOrgClusterRelationByOrgAndCluster(org.ID, int64(cluster.ID))
	if err != nil {
		return err
	}
	if relation != nil {
		return nil
	}

	relation = &db.OrgClusterRelation{
		OrgID:       req.OrgID,
		OrgName:     req.OrgName,
		ClusterID:   uint64(cluster.ID),
		ClusterName: req.ClusterName,
		Creator:     userID,
	}
	return p.dbClient.CreateOrgClusterRelation(relation)
}

// FetchOrgResources 获取企业资源情况
//
// Deprecated: The calculation caliber of the quota was reset on erda 1.3.3, please see FetchOrgClusterResource
func (p *provider) FetchOrgResources(orgID uint64) (*apistructs.OrgResourceInfo, error) {
	relations, err := p.dbClient.GetOrgClusterRelationsByOrg(int64(orgID))
	if err != nil {
		return nil, err
	}

	var (
		totalCpu float64
		totalMem float64
		usedCpu  float64
		usedMem  float64
	)
	for _, v := range relations {
		// 集群维度获取集群资源总量
		clusterRes, err := p.bdl.ResourceInfo(v.ClusterName, true)
		if err != nil {
			return nil, err
		}
		for _, node := range clusterRes.Nodes {
			cpuovercommit := 1.0
			memovercommit := 1.0

			if !node.IgnoreLabels { // dcos & edas 情况下忽略 label 过滤
				if !strutil.Exist(node.Labels, fmt.Sprintf("%s/org-%s", apistructs.PlatformLabelPrefix, v.OrgName)) {
					continue
				}
				if !strutil.Exist(node.Labels, fmt.Sprintf("%s/%s", apistructs.PlatformLabelPrefix, apistructs.StatelessLabel)) &&
					!strutil.Exist(node.Labels, fmt.Sprintf("%s/%s", apistructs.PlatformLabelPrefix, apistructs.StatefulLabel)) {
					continue
				}
				if strutil.Contains(strutil.Concat(node.Labels...), fmt.Sprintf("%s/workspace-dev", apistructs.PlatformLabelPrefix)) {
					if clusterRes.DevCPUOverCommit > cpuovercommit {
						cpuovercommit = clusterRes.DevCPUOverCommit
					}
					if clusterRes.DevMEMOverCommit > memovercommit {
						memovercommit = clusterRes.DevMEMOverCommit
					}
				}
				if strutil.Contains(strutil.Concat(node.Labels...), fmt.Sprintf("%s/workspace-test", apistructs.PlatformLabelPrefix)) {
					if clusterRes.TestCPUOverCommit > cpuovercommit {
						cpuovercommit = clusterRes.TestCPUOverCommit
					}
					if clusterRes.TestMEMOverCommit > memovercommit {
						memovercommit = clusterRes.TestMEMOverCommit
					}
				}
				if strutil.Contains(strutil.Concat(node.Labels...), fmt.Sprintf("%s/workspace-staging", apistructs.PlatformLabelPrefix)) {
					if clusterRes.StagingCPUOverCommit > cpuovercommit {
						cpuovercommit = clusterRes.StagingCPUOverCommit
					}
					if clusterRes.StagingMEMOverCommit > memovercommit {
						memovercommit = clusterRes.StagingMEMOverCommit
					}
				}
				if strutil.Contains(strutil.Concat(node.Labels...), fmt.Sprintf("%s/workspace-prod", apistructs.PlatformLabelPrefix)) {
					if clusterRes.ProdCPUOverCommit > cpuovercommit {
						cpuovercommit = clusterRes.ProdCPUOverCommit
					}
					if clusterRes.ProdMEMOverCommit > memovercommit {
						memovercommit = clusterRes.ProdMEMOverCommit
					}
				}
			}
			totalCpu += node.CPUAllocatable * cpuovercommit
			totalMem += float64(node.MemAllocatable) * memovercommit / apistructs.GB
		}
	}

	// 获取当前企业已分配资源
	projects, err := p.dbClient.ListProjectByOrgID(orgID)
	if err != nil {
		return nil, err
	}
	for _, v := range projects {
		usedCpu += v.CpuQuota
		usedMem += v.MemQuota
	}

	return &apistructs.OrgResourceInfo{
		TotalCpu:     numeral.Round(totalCpu, 2),
		TotalMem:     numeral.Round(totalMem, 2),
		AvailableCpu: numeral.Round(totalCpu-usedCpu, 2),
		AvailableMem: numeral.Round(totalMem-usedMem, 2),
	}, nil
}

// ListAllOrgClusterRelation 获取所有企业对应集群关系
func (p *provider) ListAllOrgClusterRelation() ([]db.OrgClusterRelation, error) {
	return p.dbClient.ListAllOrgClusterRelations()
}

// getOrgClusterRelationsByOrg returns the list of clusters in the organization
func (p *provider) getOrgClusterRelationsByOrg(orgID uint64) ([]db.OrgClusterRelation, error) {
	return p.dbClient.GetOrgClusterRelationsByOrg(int64(orgID))
}

func (p *provider) checkReceiveTaskRuntimeEventParam(req *apistructs.PipelineTaskEvent) error {
	if req.OrgID == "" {
		return errors.Errorf("OrgID is empty")
	}

	if req.Content.RuntimeID == "" {
		return errors.Errorf("RuntimeID is empty")
	}

	if req.Content.PipelineTaskID == 0 {
		return errors.Errorf("pipelineTaskID is empty")
	}

	return nil
}

// setReleaseCrossCluster 设置企业是否允许跨集群部署开关
func (p *provider) setReleaseCrossCluster(orgID uint64, enable bool) error {
	// 检查待更新的org是否存在
	org, err := p.dbClient.GetOrg(int64(orgID))
	if err != nil {
		return apierrors.ErrSetReleaseCrossCluster.InvalidParameter(err)
	}
	org.Config.EnableReleaseCrossCluster = enable
	return p.dbClient.DB.Model(&db.Org{}).Update(org).Error
}

// genVerifyCode 生成邀请成员加入企业的验证码
func (p *provider) genVerifyCode(identityInfo *commonpb.IdentityInfo, orgID uint64) (string, error) {
	now := time.Now()
	key := apistructs.OrgInviteCodeRedisKey.GetKey(now.Day(), identityInfo.UserID, strconv.FormatUint(orgID, 10))
	code, err := p.RedisCli.DB().Get(key).Result()
	if err == redis.Nil {
		newCode := strutil.RandStr(5) + apistructs.CodeUserID(identityInfo.UserID)
		tommory := now.AddDate(0, 0, 1).Format("2006-01-02") + " 01:00:00"
		tommoryTime, _ := time.ParseInLocation("2006-01-02 15:04:05", tommory, time.Local)
		_, err := p.RedisCli.DB().Set(key, newCode, tommoryTime.Sub(now)).Result()
		if err != nil {
			return "", err
		}

		return newCode, nil
	} else if err != nil {
		return "", err
	}

	return code, nil
}

// SetNotifyConfig 设置通知配置
func (p *provider) setNotifyConfig(orgID int64, notifyConfig *pb.SetNotifyConfigRequest) error {
	org, err := p.dbClient.GetOrg(orgID)
	if err != nil {
		return err
	}

	// 目前只开放配置语音短信通知的开关，之后会开发语音短信通知的其他配置
	org.Config.EnableMS = notifyConfig.Config.EnableMS

	return p.dbClient.UpdateOrg(&org)
}

// GetNotifyConfig 获取通知配置
func (p *provider) getNotifyConfig(orgID int64) (*pb.OrgConfig, error) {
	org, err := p.dbClient.GetOrg(orgID)
	if err != nil {
		return nil, err
	}

	return &pb.OrgConfig{
		EnableMS: org.Config.EnableMS,
		// SMTPHost:                   org.Config.SMTPHost,
		// SMTPUser:                   org.Config.SMTPUser,
		// SMTPPassword:               org.Config.SMTPPassword,
		// SMTPPort:                   org.Config.SMTPPort,
		// SMTPIsSSL:                  org.Config.SMTPIsSSL,
		// SMSKeyID:                   org.Config.SMSKeyID,
		// SMSKeySecret:               org.Config.SMSKeySecret,
		// SMSSignName:                org.Config.SMSSignName,
		// SMSMonitorTemplateCode:     org.Config.SMSMonitorTemplateCode, // 监控单独的短信模版
		// VMSKeyID:                   org.Config.VMSKeyID,
		// VMSKeySecret:               org.Config.VMSKeySecret,
		// VMSMonitorTtsCode:          org.Config.VMSMonitorTtsCode, // 监控单独的语音模版
		// VMSMonitorCalledShowNumber: org.Config.VMSMonitorCalledShowNumber,
	}, nil
}

// DereferenceCluster 解除关联集群关系
func (p *provider) dereferenceCluster(userID string, req *apistructs.DereferenceClusterRequest) error {
	clusterInfo, err := p.bdl.GetCluster(req.Cluster)
	if err != nil {
		return err
	}
	if clusterInfo == nil {
		return errors.Errorf("不存在的集群%s", req.Cluster)
	}
	if err := p.dbClient.DeleteOrgClusterRelationByClusterAndOrg(req.Cluster, req.OrgID); err != nil {
		return err
	}

	return nil
}

func getAuditMessage(org db.Org, req *pb.UpdateOrgRequest) *pb.AuditMessage {
	var messageZH, messageEN strings.Builder
	if org.DisplayName != req.DisplayName {
		messageZH.WriteString(fmt.Sprintf("组织名称由 %s 改为 %s ", org.DisplayName, req.DisplayName))
		messageEN.WriteString(fmt.Sprintf("org name updated from %s to %s ", org.DisplayName, req.DisplayName))
	}
	if org.Locale != req.Locale {
		messageZH.WriteString(fmt.Sprintf("通知语言改为%s ", func() string {
			switch req.Locale {
			case "en-US":
				return "英文"
			case "zh-CN":
				return "中文"
			default:
				return ""
			}
		}()))
		messageEN.WriteString(fmt.Sprintf("language updated to %s ", req.Locale))
	}
	if org.IsPublic != req.IsPublic {
		messageZH.WriteString(func() string {
			if req.IsPublic {
				return "改为公开组织 "
			}
			return "改为私有组织 "
		}())
		messageEN.WriteString(func() string {
			if req.IsPublic {
				return "org updated to public "
			}
			return "org updated to private "
		}())
	}
	if org.Logo != req.Logo {
		messageZH.WriteString("组织Logo发生变更 ")
		messageEN.WriteString("org Logo changed ")
	}
	if org.Desc != req.Desc {
		messageZH.WriteString("组织描述信息发生变更 ")
		messageEN.WriteString("org desc changed ")
	}
	if req.BlockoutConfig != nil {
		if org.BlockoutConfig.BlockDEV != req.BlockoutConfig.BlockDev {
			messageZH.WriteString(func() string {
				if req.BlockoutConfig.BlockDev {
					return "开发环境开启封网 "
				}
				return "开发环境关闭封网 "
			}())
			messageEN.WriteString(func() string {
				if req.BlockoutConfig.BlockDev {
					return "block network opened in dev environment "
				}
				return "block network closed in dev environment "
			}())
		}
		if org.BlockoutConfig.BlockTEST != req.BlockoutConfig.BlockTest {
			messageZH.WriteString(func() string {
				if req.BlockoutConfig.BlockTest {
					return "测试环境开启封网 "
				}
				return "测试环境关闭封网 "
			}())
			messageEN.WriteString(func() string {
				if req.BlockoutConfig.BlockTest {
					return "block network opened in test environment "
				}
				return "block network closed in test environment "
			}())
		}
		if org.BlockoutConfig.BlockStage != req.BlockoutConfig.BlockStage {
			messageZH.WriteString(func() string {
				if req.BlockoutConfig.BlockStage {
					return "预发环境开启封网 "
				}
				return "预发环境关闭封网 "
			}())
			messageEN.WriteString(func() string {
				if req.BlockoutConfig.BlockStage {
					return "block network opened in staging environment "
				}
				return "block network closed in staging environment "
			}())
		}
		if org.BlockoutConfig.BlockProd != req.BlockoutConfig.BlockProd {
			messageZH.WriteString(func() string {
				if req.BlockoutConfig.BlockProd {
					return "生产环境开启封网 "
				}
				return "生产环境关闭封网 "
			}())
			messageEN.WriteString(func() string {
				if req.BlockoutConfig.BlockProd {
					return "block network opened in prod environment "
				}
				return "block network closed in prod environment "
			}())
		}
	}
	if messageZH.Len() == 0 {
		messageZH.WriteString("无信息变更")
		messageEN.WriteString("no message changed")
	}
	return &pb.AuditMessage{
		MessageZH: messageZH.String(),
		MessageEN: messageEN.String(),
	}
}

func (p *provider) coverOrgsToDto(ctx context.Context, orgs []db.Org) ([]*pb.Org, error) {
	if len(orgs) == 0 {
		return nil, nil
	}
	var currentOrgID int64
	var err error
	v := apis.GetOrgID(ctx)
	if v != "" {
		// ignore convert error
		currentOrgID, err = strutil.Atoi64(v)
		if err != nil {
			return nil, err
		}
	}
	if currentOrgID == 0 {
		// compatible TODO: refactor it
		currentOrgID, _ = p.GetCurrentOrgByUser(apis.GetUserID(ctx))
	}

	// 封装成API所需格式
	orgDTOs := make([]*pb.Org, 0, len(orgs))
	var selected bool // 是否已选中某企业flag
	for _, org := range orgs {
		orgDTO := p.convertToOrgDTO(&org)
		if (currentOrgID == 0 || currentOrgID == org.ID) && !selected {
			orgDTO.Selected = true
			selected = true
		}
		HidePassword(orgDTO)
		orgDTOs = append(orgDTOs, orgDTO)
	}
	if !selected && len(orgDTOs) > 0 { // 用户选中某个企业后被踢出此企业后
		orgDTOs[0].Selected = true
	}

	return orgDTOs, nil
}

func HidePassword(org *pb.Org) {
	if org.Config != nil {
		org.Config.SmtpPassword = SECRECT_PLACEHOLDER
		org.Config.SmsKeySecret = SECRECT_PLACEHOLDER
	}
}

func (p *provider) convertToOrgDTO(org *db.Org, domains ...string) *pb.Org {
	if org == nil {
		return nil
	}
	domain := ""
	if len(domains) > 0 {
		domain = domains[0]
	}
	domainAndPort := strutil.Split(domain, ":", true)
	port := ""
	if len(domainAndPort) > 1 {
		port = domainAndPort[1]
	}
	concatDomain := p.Cfg.UIDomain
	if port != "" {
		concatDomain = strutil.Concat(conf.UIDomain(), ":", port)
	}

	orgDto := &pb.Org{
		ID:          uint64(org.ID),
		Name:        org.Name,
		Desc:        org.Desc,
		Logo:        filehelper.APIFileUrlRetriever(org.Logo),
		Locale:      org.Locale,
		Domain:      concatDomain,
		Creator:     org.UserID,
		OpenFdp:     org.OpenFdp,
		DisplayName: org.DisplayName,
		Type:        org.Type,
		Config: &pb.OrgConfig{
			EnablePersonalMessageEmail: org.Config.EnablePersonalMessageEmail,
			EnableMS:                   org.Config.EnableMS,
			SmtpHost:                   org.Config.SMTPHost,
			SmtpUser:                   org.Config.SMTPUser,
			SmtpPassword:               org.Config.SMTPPassword,
			SmtpPort:                   org.Config.SMTPPort,
			SmtpIsSSL:                  org.Config.SMTPIsSSL,
			SmsKeyID:                   org.Config.SMSKeyID,
			SmsKeySecret:               org.Config.SMSKeySecret,
			SmsSignName:                org.Config.SMSSignName,
			SmsMonitorTemplateCode:     org.Config.SMSMonitorTemplateCode, // 监控单独的短信模版
			VmsKeyID:                   org.Config.VMSKeyID,
			VmsKeySecret:               org.Config.VMSKeySecret,
			VmsMonitorTtsCode:          org.Config.VMSMonitorTtsCode, // 监控单独的语音模版
			VmsMonitorCalledShowNumber: org.Config.VMSMonitorCalledShowNumber,
			EnableAI:                   org.Config.EnableAI,
		},
		IsPublic: org.IsPublic,
		BlockoutConfig: &pb.BlockoutConfig{
			BlockDev:   org.BlockoutConfig.BlockDEV,
			BlockTest:  org.BlockoutConfig.BlockTEST,
			BlockStage: org.BlockoutConfig.BlockStage,
			BlockProd:  org.BlockoutConfig.BlockProd,
		},
		EnableReleaseCrossCluster: org.Config.EnableReleaseCrossCluster,
		CreatedAt:                 timestamppb.New(org.CreatedAt),
		UpdatedAt:                 timestamppb.New(org.UpdatedAt),
	}
	if orgDto.DisplayName == "" {
		orgDto.DisplayName = orgDto.Name
	}
	return orgDto
}
