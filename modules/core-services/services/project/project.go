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

// Package project 封装项目资源相关操作
package project

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/providers/i18n"
	dashboardPb "github.com/erda-project/erda-proto-go/cmp/dashboard/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/core-services/conf"
	"github.com/erda-project/erda/modules/core-services/dao"
	"github.com/erda-project/erda/modules/core-services/model"
	"github.com/erda-project/erda/modules/core-services/types"
	"github.com/erda-project/erda/pkg/crypto/uuid"
	"github.com/erda-project/erda/pkg/filehelper"
	"github.com/erda-project/erda/pkg/numeral"
	calcu "github.com/erda-project/erda/pkg/resourcecalculator"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/erda-project/erda/pkg/ucauth"
)

// Project 资源对象操作封装
type Project struct {
	db    *dao.DBClient
	uc    *ucauth.UCClient
	bdl   *bundle.Bundle
	trans i18n.Translator

	clusterResourceClient dashboardPb.ClusterResourceServer
}

// Option 定义 Project 对象的配置选项
type Option func(*Project)

// New 新建 Project 实例，通过 Project 实例操作企业资源
func New(options ...Option) *Project {
	project := &Project{}
	for _, f := range options {
		f(project)
	}
	return project
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(project *Project) {
		project.db = db
	}
}

// WithUCClient 配置 uc client
func WithUCClient(uc *ucauth.UCClient) Option {
	return func(project *Project) {
		project.uc = uc
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(project *Project) {
		project.bdl = bdl
	}
}

// WithClusterResourceClient set the gRPC client of CMP cluster resource
func WithClusterResourceClient(cli dashboardPb.ClusterResourceServer) Option {
	return func(project *Project) {
		project.clusterResourceClient = cli
	}
}

// WithI18n set the translator
func WithI18n(translator i18n.Translator) Option {
	return func(project *Project) {
		project.trans = translator
	}
}

// CreateWithEvent 创建项目 & 发送事件
func (p *Project) CreateWithEvent(userID string, createReq *apistructs.ProjectCreateRequest) (int64, error) {
	// 创建项目
	if createReq.DisplayName == "" {
		createReq.DisplayName = createReq.Name
	}
	project, err := p.Create(userID, createReq)
	if err != nil {
		return 0, err
	}
	ev := &apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			Event:         bundle.ProjectEvent,
			Action:        bundle.CreateAction,
			OrgID:         strconv.FormatInt(project.OrgID, 10),
			ProjectID:     strconv.FormatInt(project.ID, 10),
			ApplicationID: "-1",
			TimeStamp:     time.Now().Format("2006-01-02 15:04:05"),
		},
		Sender:  bundle.SenderCoreServices,
		Content: *project,
	}
	// 发送项目创建事件
	if err = p.bdl.CreateEvent(ev); err != nil {
		logrus.Warnf("failed to send project create event, (%v)", err)
	}

	return project.ID, nil
}

// Create 创建项目
func (p *Project) Create(userID string, createReq *apistructs.ProjectCreateRequest) (*model.Project, error) {
	// 参数合法性检查
	if createReq.Name == "" {
		return nil, errors.Errorf("failed to create project(name is empty)")
	}
	if createReq.OrgID == 0 {
		return nil, errors.Errorf("failed to create project(org id is empty)")
	}
	// 只有 DevOps 类型的项目，才能配置 quota
	if createReq.Template != apistructs.DevopsTemplate {
		createReq.ResourceConfigs = nil
	}
	var clusterConfig []byte
	if createReq.ResourceConfigs != nil {
		if err := createReq.ResourceConfigs.Check(); err != nil {
			return nil, err
		}
		createReq.ClusterConfig = map[string]string{
			"PROD":    createReq.ResourceConfigs.PROD.ClusterName,
			"STAGING": createReq.ResourceConfigs.STAGING.ClusterName,
			"TEST":    createReq.ResourceConfigs.TEST.ClusterName,
			"DEV":     createReq.ResourceConfigs.DEV.ClusterName,
		}
		clusterConfig, _ = json.Marshal(createReq.ClusterConfig)
	}

	if err := initRollbackConfig(&createReq.RollbackConfig); err != nil {
		return nil, err
	}
	rollbackConfig, err := json.Marshal(createReq.RollbackConfig)
	if err != nil {
		logrus.Infof("failed to marshal rollbackConfig, (%v)", err)
		return nil, errors.Errorf("failed to marshal rollbackConfig")
	}

	functions, err := json.Marshal(createReq.Template.GetProjectFunctionsByTemplate())
	if err != nil {
		logrus.Infof("failed to marshal projectFunctions, (%v)", err)
		return nil, errors.Errorf("failed to marshal projectFunctions")
	}

	project, err := p.db.GetProjectByOrgAndName(int64(createReq.OrgID), createReq.Name)
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
	}
	if project != nil {
		return nil, errors.Errorf("failed to create project(name already exists)")
	}

	tx := p.db.Begin()
	defer tx.RollbackUnlessCommitted()

	now := time.Now()
	// 添加项目至DB
	project = &model.Project{
		BaseModel:      model.BaseModel{},
		Name:           createReq.Name,
		DisplayName:    createReq.DisplayName,
		Desc:           createReq.Desc,
		Logo:           createReq.Logo,
		OrgID:          int64(createReq.OrgID),
		UserID:         userID,
		DDHook:         createReq.DdHook,
		ClusterConfig:  string(clusterConfig),
		RollbackConfig: string(rollbackConfig),
		CpuQuota:       createReq.CpuQuota,
		MemQuota:       createReq.MemQuota,
		Functions:      string(functions),
		ActiveTime:     now,
		EnableNS:       conf.EnableNS(),
		IsPublic:       false,
		Type:           string(createReq.Template),
	}
	if err := tx.Create(&project).Error; err != nil {
		logrus.WithError(err).WithField("model", project.TableName()).
			Errorln("failed to Create")
		return nil, errors.Errorf("failed to insert project to database")
	}

	// record quota if it is configured
	logrus.WithField("createReq.ResourceConfigs", createReq.ResourceConfigs).Infoln()
	if createReq.ResourceConfigs != nil {
		quota := &model.ProjectQuota{
			ProjectID:          uint64(project.ID),
			ProjectName:        createReq.Name,
			ProdClusterName:    createReq.ResourceConfigs.PROD.ClusterName,
			StagingClusterName: createReq.ResourceConfigs.STAGING.ClusterName,
			TestClusterName:    createReq.ResourceConfigs.TEST.ClusterName,
			DevClusterName:     createReq.ResourceConfigs.DEV.ClusterName,
			ProdCPUQuota:       calcu.CoreToMillcore(createReq.ResourceConfigs.PROD.CPUQuota),
			ProdMemQuota:       calcu.GibibyteToByte(createReq.ResourceConfigs.PROD.MemQuota),
			StagingCPUQuota:    calcu.CoreToMillcore(createReq.ResourceConfigs.STAGING.CPUQuota),
			StagingMemQuota:    calcu.GibibyteToByte(createReq.ResourceConfigs.STAGING.MemQuota),
			TestCPUQuota:       calcu.CoreToMillcore(createReq.ResourceConfigs.TEST.CPUQuota),
			TestMemQuota:       calcu.GibibyteToByte(createReq.ResourceConfigs.TEST.MemQuota),
			DevCPUQuota:        calcu.CoreToMillcore(createReq.ResourceConfigs.DEV.CPUQuota),
			DevMemQuota:        calcu.GibibyteToByte(createReq.ResourceConfigs.DEV.MemQuota),
			CreatorID:          userID,
			UpdaterID:          userID,
		}
		if err := tx.Debug().Create(&quota).Error; err != nil {
			logrus.WithError(err).WithField("model", quota.TableName()).
				Errorln("failed to Create")
			return nil, errors.Errorf("failed to insert project quota to database")
		}
	}

	// 新增项目管理员至admin_members表
	users, err := p.uc.FindUsers([]string{userID})
	if err != nil {
		logrus.Warnf("user query error: %v", err)
	}
	if len(users) > 0 {
		member := model.Member{
			ScopeType:  apistructs.ProjectScope,
			ScopeID:    project.ID,
			ScopeName:  project.Name,
			ParentID:   project.OrgID,
			UserID:     userID,
			Email:      users[0].Email,
			Mobile:     users[0].Phone,
			Name:       users[0].Name,
			Nick:       users[0].Nick,
			Avatar:     users[0].AvatarURL,
			UserSyncAt: time.Now(),
			OrgID:      project.OrgID,
			ProjectID:  project.ID,
			Token:      uuid.UUID(),
		}
		memberExtra := model.MemberExtra{
			UserID:        userID,
			ScopeID:       project.ID,
			ScopeType:     apistructs.ProjectScope,
			ParentID:      project.OrgID,
			ResourceKey:   apistructs.RoleResourceKey,
			ResourceValue: types.RoleProjectOwner,
		}
		if err := tx.Create(&member).Error; err != nil {
			logrus.WithError(err).WithField("model", member.TableName()).
				Errorln("failed to add member to database while creating project")
			return nil, errors.Errorf("failed to add member to database while creating project")
		}
		if err := tx.Create(&memberExtra).Error; err != nil {
			logrus.WithError(err).WithField("model", memberExtra.TableName()).
				Errorln("failed to add member roles to database while creating project")
			return nil, errors.Errorf("failed to add member roles to database while creating project")
		}
	}

	tx.Commit()

	return project, nil
}

// UpdateWithEvent 更新项目 & 发送事件
func (p *Project) UpdateWithEvent(orgID, projectID int64, userID string, updateReq *apistructs.ProjectUpdateBody) error {
	// 更新项目
	project, err := p.Update(orgID, projectID, userID, updateReq)
	if err != nil {
		return err
	}

	ev := &apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			Event:         bundle.ProjectEvent,
			Action:        bundle.UpdateAction,
			OrgID:         strconv.FormatInt(project.OrgID, 10),
			ProjectID:     strconv.FormatInt(project.ID, 10),
			ApplicationID: "-1",
			TimeStamp:     time.Now().Format("2006-01-02 15:04:05"),
		},
		Sender:  bundle.SenderCoreServices,
		Content: *project,
	}
	// 发送项目更新事件
	if err = p.bdl.CreateEvent(ev); err != nil {
		logrus.Warnf("failed to send project update event, (%v)", err)
	}

	return nil
}

// Update 更新项目
func (p *Project) Update(orgID, projectID int64, userID string, updateReq *apistructs.ProjectUpdateBody) (*model.Project, error) {
	data, _ := json.Marshal(updateReq)
	logrus.Infof("updateReq: %s", string(data))
	if updateReq.ResourceConfigs != nil {
		updateReq.ClusterConfig = map[string]string{
			"PROD":    updateReq.ResourceConfigs.PROD.ClusterName,
			"STAGING": updateReq.ResourceConfigs.STAGING.ClusterName,
			"TEST":    updateReq.ResourceConfigs.TEST.ClusterName,
			"DEV":     updateReq.ResourceConfigs.DEV.ClusterName,
		}
		if err := updateReq.ResourceConfigs.Check(); err != nil {
			return nil, err
		}
	}
	if err := checkRollbackConfig(&updateReq.RollbackConfig); err != nil {
		return nil, err
	}

	// 检查待更新的project是否存在
	project, err := p.db.GetProjectByID(projectID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update project")
	}

	if err := patchProject(&project, updateReq); err != nil {
		return nil, err
	}

	tx := p.db.Begin()
	defer tx.RollbackUnlessCommitted()

	if err = tx.Save(&project).Error; err != nil {
		logrus.Warnf("failed to update project, (%v)", err)
		return nil, errors.Errorf("failed to update project")
	}

	var oldQuota = new(model.ProjectQuota)
	err = p.db.First(oldQuota, map[string]interface{}{"project_id": projectID}).Error
	hasOldQuota := err == nil

	if updateReq.ResourceConfigs == nil {
		if hasOldQuota {
			return nil, errors.Errorf("cant not update project quota to empty")
		}
		tx.Commit()
		return &project, nil
	}

	// create or update quota
	var quota = model.ProjectQuota{
		ProjectID:          uint64(projectID),
		ProjectName:        updateReq.Name,
		ProdClusterName:    updateReq.ResourceConfigs.PROD.ClusterName,
		StagingClusterName: updateReq.ResourceConfigs.STAGING.ClusterName,
		TestClusterName:    updateReq.ResourceConfigs.TEST.ClusterName,
		DevClusterName:     updateReq.ResourceConfigs.DEV.ClusterName,
		ProdCPUQuota:       calcu.CoreToMillcore(updateReq.ResourceConfigs.PROD.CPUQuota),
		ProdMemQuota:       calcu.GibibyteToByte(updateReq.ResourceConfigs.PROD.MemQuota),
		StagingCPUQuota:    calcu.CoreToMillcore(updateReq.ResourceConfigs.PROD.CPUQuota),
		StagingMemQuota:    calcu.GibibyteToByte(updateReq.ResourceConfigs.STAGING.MemQuota),
		TestCPUQuota:       calcu.CoreToMillcore(updateReq.ResourceConfigs.TEST.CPUQuota),
		TestMemQuota:       calcu.GibibyteToByte(updateReq.ResourceConfigs.TEST.MemQuota),
		DevCPUQuota:        calcu.CoreToMillcore(updateReq.ResourceConfigs.DEV.CPUQuota),
		DevMemQuota:        calcu.GibibyteToByte(updateReq.ResourceConfigs.DEV.MemQuota),
		CreatorID:          userID,
		UpdaterID:          userID,
	}
	if hasOldQuota {
		quota.ID = oldQuota.ID
		quota.CreatorID = oldQuota.CreatorID
		err = tx.Debug().Save(&quota).Error
	} else {
		err = tx.Debug().Create(&quota).Error
	}
	if err != nil {
		logrus.WithError(err).Errorln("failed to update project quota")
		return nil, errors.Errorf("failed to update project quota: %v", err)
	}
	if err = tx.Commit().Error; err != nil {
		err = errors.Wrap(err, "failed commit to update project and quota")
		logrus.WithError(err).Errorln()
		return nil, err
	}

	// audit
	go func() {
		if !isQuotaChanged(*oldQuota, quota) {
			return
		}
		var orgName = strconv.FormatInt(orgID, 10)
		if org, err := p.db.GetOrg(orgID); err == nil {
			orgName = fmt.Sprintf("%s(%s)", org.Name, org.DisplayName)
		}
		auditCtx := map[string]interface{}{
			"orgName":       orgName,
			"projectName":   project.Name,
			"newDevCPU":     calcu.ResourceToString(float64(quota.DevCPUQuota), "cpu"),
			"newDevMem":     calcu.ResourceToString(float64(quota.DevMemQuota), "memory"),
			"newTestCPU":    calcu.ResourceToString(float64(quota.TestCPUQuota), "cpu"),
			"newTestMem":    calcu.ResourceToString(float64(quota.TestMemQuota), "memory"),
			"newStagingCPU": calcu.ResourceToString(float64(quota.StagingCPUQuota), "cpu"),
			"newStagingMem": calcu.ResourceToString(float64(quota.StagingMemQuota), "memory"),
			"newProdCPU":    calcu.ResourceToString(float64(quota.ProdCPUQuota), "cpu"),
			"newProdMem":    calcu.ResourceToString(float64(quota.ProdMemQuota), "memory"),
			"oldDevCPU":     calcu.ResourceToString(float64(oldQuota.DevCPUQuota), "cpu"),
			"oldDevMem":     calcu.ResourceToString(float64(oldQuota.DevMemQuota), "memory"),
			"oldTestCPU":    calcu.ResourceToString(float64(oldQuota.TestCPUQuota), "cpu"),
			"oldTestMem":    calcu.ResourceToString(float64(oldQuota.TestMemQuota), "memory"),
			"oldStagingCPU": calcu.ResourceToString(float64(oldQuota.StagingCPUQuota), "cpu"),
			"oldStagingMem": calcu.ResourceToString(float64(oldQuota.StagingMemQuota), "memory"),
			"oldProdCPU":    calcu.ResourceToString(float64(oldQuota.ProdCPUQuota), "cpu"),
			"oldProdMem":    calcu.ResourceToString(float64(oldQuota.ProdMemQuota), "memory"),
		}
		now := time.Now().Format("2006-01-02 15:04:05")
		audit := apistructs.Audit{
			UserID:       userID,
			ScopeType:    apistructs.OrgScope,
			ScopeID:      uint64(orgID),
			OrgID:        uint64(orgID),
			ProjectID:    uint64(projectID),
			Context:      auditCtx,
			TemplateName: "updateProjectQuota",
			Result:       "success",
			StartTime:    now,
			EndTime:      now,
		}
		auditRecord, err := convertAuditCreateReq2Model(audit)
		if err != nil {
			logrus.WithError(err).WithField("audit", audit).
				Errorf("failed to convertAuditCreateReq2Model")
			return
		}
		if err = p.db.CreateAudit(auditRecord); err != nil {
			logrus.Errorf("failed to create quota audit event when update project %s, %v", project.Name, err)
		}
	}()

	return &project, nil
}

func isQuotaChanged(oldQuota, newQuota model.ProjectQuota) bool {
	if oldQuota.DevCPUQuota != newQuota.DevCPUQuota || oldQuota.DevMemQuota != newQuota.DevMemQuota ||
		oldQuota.TestCPUQuota != newQuota.TestCPUQuota || oldQuota.TestMemQuota != newQuota.TestMemQuota ||
		oldQuota.StagingCPUQuota != newQuota.StagingCPUQuota || oldQuota.StagingMemQuota != newQuota.StagingMemQuota ||
		oldQuota.ProdCPUQuota != newQuota.ProdCPUQuota || oldQuota.ProdMemQuota != newQuota.ProdMemQuota {
		return true
	}
	return false
}

func patchProject(project *model.Project, updateReq *apistructs.ProjectUpdateBody) error {
	clusterConfig, err := json.Marshal(updateReq.ResourceConfigs)
	if err != nil {
		logrus.Errorf("failed to marshal clusterConfig, (%v)", err)
		return errors.Errorf("failed to marshal clusterConfig")
	}

	rollbackConfig, err := json.Marshal(updateReq.RollbackConfig)
	if err != nil {
		logrus.Errorf("failed to marshal rollbackConfig, (%v)", err)
		return errors.Errorf("failed to marshal rollbackConfig")
	}

	if updateReq.DisplayName != "" {
		project.DisplayName = updateReq.DisplayName
	}

	if updateReq.ResourceConfigs != nil {
		project.ClusterConfig = string(clusterConfig)
	}

	if len(updateReq.RollbackConfig) != 0 {
		project.RollbackConfig = string(rollbackConfig)
	}

	project.Desc = updateReq.Desc
	project.Logo = updateReq.Logo
	project.DDHook = updateReq.DdHook
	project.CpuQuota = updateReq.CpuQuota
	project.MemQuota = updateReq.MemQuota
	project.ActiveTime = time.Now()
	project.IsPublic = updateReq.IsPublic

	return nil
}

func convertAuditCreateReq2Model(req apistructs.Audit) (*model.Audit, error) {
	context, err := json.Marshal(req.Context)
	if err != nil {
		return nil, err
	}
	startAt, err := time.ParseInLocation("2006-01-02 15:04:05", req.StartTime, time.Local)
	if err != nil {
		return nil, err
	}
	endAt, err := time.ParseInLocation("2006-01-02 15:04:05", req.EndTime, time.Local)
	if err != nil {
		return nil, err
	}
	audit := &model.Audit{
		StartTime:    startAt,
		EndTime:      endAt,
		UserID:       req.UserID,
		ScopeType:    req.ScopeType,
		ScopeID:      req.ScopeID,
		FDPProjectID: req.FDPProjectID,
		AppID:        req.AppID,
		OrgID:        req.OrgID,
		ProjectID:    req.ProjectID,
		Context:      string(context),
		TemplateName: req.TemplateName,
		AuditLevel:   req.AuditLevel,
		Result:       req.Result,
		ErrorMsg:     req.ErrorMsg,
		ClientIP:     req.ClientIP,
		UserAgent:    req.UserAgent,
	}

	return audit, nil
}

// DeleteWithEvent 删除项目 & 发送事件
func (p *Project) DeleteWithEvent(projectID int64) error {
	// 删除项目
	project, err := p.Delete(projectID)
	if err != nil {
		return err
	}

	ev := &apistructs.EventCreateRequest{
		EventHeader: apistructs.EventHeader{
			Event:         bundle.ProjectEvent,
			Action:        bundle.DeleteAction,
			OrgID:         strconv.FormatInt(project.OrgID, 10),
			ProjectID:     strconv.FormatInt(projectID, 10),
			ApplicationID: "-1",
			TimeStamp:     time.Now().Format("2006-01-02 15:04:05"),
		},
		Sender:  bundle.SenderCoreServices,
		Content: *project,
	}
	// 发送项目删除事件
	if err = p.bdl.CreateEvent(ev); err != nil {
		logrus.Warnf("failed to send project delete event, (%v)", err)
	}

	return nil
}

// Delete 删除项目
func (p *Project) Delete(projectID int64) (*model.Project, error) {
	// check if application exists
	if count, err := p.db.GetApplicationCountByProjectID(projectID); err != nil || count > 0 {
		return nil, errors.Errorf("failed to delete project(there exists applications)")
	}

	project, err := p.db.GetProjectByID(projectID)
	if err != nil {
		return nil, errors.Errorf("failed to get project, (%v)", err)
	}

	// TODO We need to turn this check on after adding the delete portal to the UI
	// check if addon exists
	// addOnListResp, err := p.bdl.ListAddonByProjectID(projectID, project.OrgID)
	// if err != nil {
	// 	return nil, err
	// }
	// if addOnListResp != nil && len(addOnListResp.Data) > 0 {
	// 	return nil, errors.Errorf("failed to delete project(there exists addons)")
	// }

	if err = p.db.DeleteProject(projectID); err != nil {
		return nil, errors.Errorf("failed to delete project, (%v)", err)
	}
	_ = p.db.DeleteProjectQutoa(projectID)
	logrus.Infof("deleted project %d", projectID)

	// 删除权限表记录
	if err = p.db.DeleteMembersByScope(apistructs.ProjectScope, projectID); err != nil {
		logrus.Warnf("failed to delete members, (%v)", err)
	}
	if err = p.db.DeleteMemberExtraByScope(apistructs.ProjectScope, projectID); err != nil {
		logrus.Warnf("failed to delete members extra, (%v)", err)
	}
	return &project, nil
}

// Get 获取项目
func (p *Project) Get(ctx context.Context, projectID int64) (*apistructs.ProjectDTO, error) {
	langCodes, _ := ctx.Value("lang_codes").(i18n.LanguageCodes)

	project, err := p.db.GetProjectByID(projectID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}
	projectDTO := p.convertToProjectDTO(true, &project)

	owners, err := p.db.GetMemberByScopeAndRole(apistructs.ProjectScope, []uint64{uint64(projectID)}, []string{"owner"})
	if err != nil {
		return nil, err
	}
	for _, v := range owners {
		projectDTO.Owners = append(projectDTO.Owners, v.UserID)
	}

	// 查询项目 quota
	p.fetchQuota(&projectDTO)
	// 查询项目下的 pod 的 request 数据
	p.fetchPodInfo(&projectDTO)
	// 根据已有统计值计算比率
	p.calcuRequestRate(&projectDTO)
	// 查询对应的集群实际可用资源
	p.fetchAvailable(ctx, &projectDTO)
	// 对比 quota 和实际可用资源，打 tips
	p.makeProjectDtoTips(&projectDTO, langCodes)

	return &projectDTO, nil
}

func (p *Project) fetchQuota(dto *apistructs.ProjectDTO) {
	var projectQuota model.ProjectQuota
	if err := p.db.First(&projectQuota, map[string]interface{}{"project_id": dto.ID}).Error; err != nil {
		logrus.WithError(err).WithField("project_id", dto.ID).
			Warnln("failed to select the quota record of the project")
		return
	}
	dto.ClusterConfig = make(map[string]string)
	dto.ResourceConfig = apistructs.NewResourceConfig()
	dto.ClusterConfig["PROD"] = projectQuota.ProdClusterName
	dto.ClusterConfig["STAGING"] = projectQuota.StagingClusterName
	dto.ClusterConfig["TEST"] = projectQuota.TestClusterName
	dto.ClusterConfig["DEV"] = projectQuota.DevClusterName
	dto.ResourceConfig.PROD.ClusterName = projectQuota.ProdClusterName
	dto.ResourceConfig.STAGING.ClusterName = projectQuota.StagingClusterName
	dto.ResourceConfig.TEST.ClusterName = projectQuota.TestClusterName
	dto.ResourceConfig.DEV.ClusterName = projectQuota.DevClusterName
	dto.ResourceConfig.PROD.CPUQuota = calcu.MillcoreToCore(projectQuota.ProdCPUQuota, 3)
	dto.ResourceConfig.STAGING.CPUQuota = calcu.MillcoreToCore(projectQuota.StagingCPUQuota, 3)
	dto.ResourceConfig.TEST.CPUQuota = calcu.MillcoreToCore(projectQuota.TestCPUQuota, 3)
	dto.ResourceConfig.DEV.CPUQuota = calcu.MillcoreToCore(projectQuota.DevCPUQuota, 3)
	dto.ResourceConfig.PROD.MemQuota = calcu.ByteToGibibyte(projectQuota.ProdMemQuota, 3)
	dto.ResourceConfig.STAGING.MemQuota = calcu.ByteToGibibyte(projectQuota.StagingMemQuota, 3)
	dto.ResourceConfig.TEST.MemQuota = calcu.ByteToGibibyte(projectQuota.TestMemQuota, 3)
	dto.ResourceConfig.DEV.MemQuota = calcu.ByteToGibibyte(projectQuota.DevMemQuota, 3)
	dto.CpuQuota = calcu.MillcoreToCore(projectQuota.ProdCPUQuota+projectQuota.StagingCPUQuota+projectQuota.TestCPUQuota+projectQuota.DevCPUQuota, 3)
	dto.MemQuota = calcu.ByteToGibibyte(projectQuota.ProdMemQuota+projectQuota.StagingMemQuota+projectQuota.TestMemQuota+projectQuota.DevMemQuota, 3)
}

func (p *Project) fetchPodInfo(dto *apistructs.ProjectDTO) {
	var podInfos []apistructs.PodInfo
	if err := p.db.Find(&podInfos, map[string]interface{}{"project_id": dto.ID}).Error; err != nil {
		logrus.WithError(err).WithField("project_id", dto.ID).
			Warnln("failed to Find the namespaces info in the project")
		return
	}

	for _, podInfo := range podInfos {
		for workspace, resourceConfig := range map[string]*apistructs.ResourceConfigInfo{
			"prod":    dto.ResourceConfig.PROD,
			"staging": dto.ResourceConfig.STAGING,
			"test":    dto.ResourceConfig.TEST,
			"dev":     dto.ResourceConfig.DEV,
		} {
			if !strings.EqualFold(podInfo.Workspace, workspace) {
				continue
			}
			if podInfo.Cluster != resourceConfig.ClusterName {
				continue
			}
			resourceConfig.CPURequest += podInfo.CPURequest
			resourceConfig.MemRequest += podInfo.MemRequest / 1024
			switch podInfo.ServiceType {
			case "addon":
				resourceConfig.CPURequestByAddon += podInfo.CPURequest
				resourceConfig.MemRequestByAddon += podInfo.MemRequest / 1024
			case "stateless-service":
				resourceConfig.CPURequestByService += podInfo.CPURequest
				resourceConfig.MemRequestByService += podInfo.MemRequest / 1024
			}
		}
	}
}

// 查出各环境的实际可用资源
// 各环境的实际可用资源 = 有该环境标签的所有集群的可用资源之和
// 每台机器的可用资源 = 该机器的 allocatable - 该机器的 request
func (p *Project) fetchAvailable(ctx context.Context, dto *apistructs.ProjectDTO) {
	if dto.ResourceConfig == nil {
		return
	}
	clusterNames := strutil.DedupSlice([]string{
		dto.ResourceConfig.PROD.ClusterName,
		dto.ResourceConfig.STAGING.ClusterName,
		dto.ResourceConfig.TEST.ClusterName,
		dto.ResourceConfig.DEV.ClusterName,
	})
	req := &dashboardPb.GetClustersResourcesRequest{ClusterNames: clusterNames}
	ctx, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()
	clustersResources, err := p.clusterResourceClient.GetClustersResources(ctx, req)
	if err != nil {
		logrus.WithError(err).WithField("func", "fetchAvailable").
			Errorf("failed to GetClustersResources, clusterNames: %v", clusterNames)
		return
	}
	for _, clusterItem := range clustersResources.List {
		if !clusterItem.GetSuccess() {
			logrus.WithField("cluster_name", clusterItem.GetClusterName()).WithField("err", clusterItem.GetErr()).
				Warnln("the cluster is not valid now")
			continue
		}
		for _, host := range clusterItem.Hosts {
			for _, label := range host.Labels {
				var source *apistructs.ResourceConfigInfo
				switch strings.ToLower(label) {
				case "dice/workspace-prod=true":
					source = dto.ResourceConfig.PROD
				case "dice/workspace-staging=true":
					source = dto.ResourceConfig.STAGING
				case "dice/workspace-test=true":
					source = dto.ResourceConfig.TEST
				case "dice/workspace-dev=true":
					source = dto.ResourceConfig.DEV
				}
				if source != nil && source.ClusterName == clusterItem.GetClusterName() {
					source.CPUAvailable += calcu.MillcoreToCore(host.GetCpuAllocatable()-host.GetCpuRequest(), 3)
					source.MemAvailable += calcu.ByteToGibibyte(host.GetMemAllocatable()-host.GetMemRequest(), 3)
				}
			}
		}
	}
}

// 根据已有统计值计算比率
func (p *Project) calcuRequestRate(dto *apistructs.ProjectDTO) {
	if dto.ResourceConfig == nil {
		return
	}
	for _, source := range []*apistructs.ResourceConfigInfo{
		dto.ResourceConfig.PROD,
		dto.ResourceConfig.STAGING,
		dto.ResourceConfig.TEST,
		dto.ResourceConfig.DEV,
	} {
		if source.CPUQuota != 0 {
			source.CPURequestRate = source.CPURequest / source.CPUQuota
			source.CPURequestByAddonRate = source.CPURequestByAddon / source.CPUQuota
			source.CPURequestByServiceRate = source.CPURequestByService / source.CPUQuota
		}
		if source.MemQuota != 0 {
			source.MemRequestRate = source.MemRequest / source.MemQuota
			source.MemRequestByAddonRate = source.MemRequestByAddon / source.MemQuota
			source.MemRequestByServiceRate = source.MemRequestByService / source.MemQuota
		}
	}
}

func (p *Project) makeProjectDtoTips(dto *apistructs.ProjectDTO, langCodes i18n.LanguageCodes) {
	if dto.ResourceConfig == nil {
		return
	}
	for _, source := range []*apistructs.ResourceConfigInfo{
		dto.ResourceConfig.PROD,
		dto.ResourceConfig.STAGING,
		dto.ResourceConfig.TEST,
		dto.ResourceConfig.DEV,
	} {
		if source.CPUAvailable < source.CPUQuota || source.MemAvailable < source.MemQuota {
			source.Tips = p.trans.Text(langCodes, "AvailableIsLessThanQuota")
		}
	}
}

// GetModelProject 获取项目
func (p *Project) GetModelProject(projectID int64) (*model.Project, error) {
	project, err := p.db.GetProjectByID(projectID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	return &project, nil
}

func (p *Project) GetModelProjectsMap(projectIDs []uint64) (map[int64]*model.Project, error) {
	_, projects, err := p.db.GetProjectsByIDs(projectIDs, &apistructs.ProjectListRequest{
		PageNo:   1,
		PageSize: len(projectIDs),
	})
	if err != nil {
		return nil, errors.Errorf("failed to get projects, (%v)", err)
	}

	projectMap := make(map[int64]*model.Project)
	for i, p := range projects {
		projectMap[p.ID] = &projects[i]
	}
	return projectMap, nil
}

// FillQuota 根据项目资源使用情况填充项目资源配额
func (p *Project) FillQuota(orgResources map[uint64]apistructs.OrgResourceInfo) error {
	for k, v := range orgResources {
		projects, err := p.db.ListProjectByOrgID(k)
		if err != nil {
			return err
		}
		// 获取当前企业下各项目已使用资源
		projectIDs := make([]uint64, 0, len(projects))
		for _, proj := range projects {
			projectIDs = append(projectIDs, uint64(proj.ID))
		}
		resp, err := p.bdl.ProjectResource(projectIDs)
		if err != nil {
			return err
		}

		// 企业资源使用
		var (
			orgCpuUsed float64
			orgMemUsed float64
		)

		projectCpuUsed := make(map[uint64]float64, len(projects))
		projectMemUsed := make(map[uint64]float64, len(projects))
		for pk, pv := range resp.Data {
			projectCpuUsed[pk] = pv.CpuServiceUsed + pv.CpuAddonUsed
			projectMemUsed[pk] = pv.MemServiceUsed + pv.MemAddonUsed

			orgCpuUsed += projectCpuUsed[pk]
			orgMemUsed += projectMemUsed[pk]
		}
		for i, proj := range projects {
			// 修正已有项目无配额情况
			if math.Round(proj.CpuQuota) == 0 {
				projectCpuQuota := projectCpuUsed[uint64(proj.ID)] + projectCpuUsed[uint64(proj.ID)]*(v.TotalCpu-orgCpuUsed)/orgCpuUsed
				projects[i].CpuQuota = numeral.Round(projectCpuQuota, 2)
			}
			if math.Round(proj.MemQuota) == 0 {
				projectMemQuota := projectMemUsed[uint64(proj.ID)] + projectMemUsed[uint64(proj.ID)]*(v.TotalMem-orgMemUsed)/orgMemUsed
				projects[i].MemQuota = numeral.Round(projectMemQuota, 2)
			}
			if err := p.db.UpdateProject(&projects[i]); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetAllProjects list all project
func (p *Project) GetAllProjects() ([]apistructs.ProjectDTO, error) {
	projects, err := p.db.GetAllProjects()
	if err != nil {
		return nil, err
	}
	projectsDTO := make([]apistructs.ProjectDTO, 0, len(projects))
	for _, v := range projects {
		projectsDTO = append(projectsDTO, p.convertToProjectDTO(true, &v))
	}
	return projectsDTO, nil
}

// ListAllProjects 企业管理员可查看当前企业下所有项目，包括未加入的项目
func (p *Project) ListAllProjects(userID string, params *apistructs.ProjectListRequest) (
	*apistructs.PagingProjectDTO, error) {
	total, projects, err := p.db.GetProjectsByOrgIDAndName(int64(params.OrgID), params)
	if err != nil {
		return nil, errors.Errorf("failed to get projects, (%v)", err)
	}
	members, err := p.db.GetMembersByScopeTypeAndUser(apistructs.ProjectScope, userID)
	if err != nil {
		return nil, errors.Errorf("failed to get projects, (%v)", err)
	}

	// 转换成所需格式
	projectDTOs := make([]apistructs.ProjectDTO, 0, len(projects))
	projectIDs := make([]uint64, 0, len(projects))
	for i := range projects {
		// 找出企业管理员已加入的项目
		flag := params.Joined
		for j := range members {
			if projects[i].ID == members[j].ScopeID {
				flag = true
			}
		}
		projectDTOs = append(projectDTOs, p.convertToProjectDTO(flag, &projects[i]))
		projectIDs = append(projectIDs, uint64(projects[i].ID))
	}
	resp, err := p.bdl.ProjectResource(projectIDs)
	if err != nil {
		return nil, err
	}

	projectOwnerMap := make(map[uint64][]string)
	owners, err := p.db.GetMemberByScopeAndRole(apistructs.ProjectScope, projectIDs, []string{types.RoleProjectOwner})
	if err != nil {
		return nil, err
	}
	for _, v := range owners {
		projectID := uint64(v.ScopeID)
		projectOwnerMap[projectID] = append(projectOwnerMap[projectID], v.UserID)
	}

	for i := range projectDTOs {
		if v, ok := resp.Data[projectDTOs[i].ID]; ok {
			projectDTOs[i].CpuServiceUsed = v.CpuServiceUsed
			projectDTOs[i].MemServiceUsed = v.MemServiceUsed
			projectDTOs[i].CpuAddonUsed = v.CpuAddonUsed
			projectDTOs[i].MemAddonUsed = v.MemAddonUsed
		}
		if v, ok := projectOwnerMap[projectDTOs[i].ID]; ok {
			projectDTOs[i].Owners = v
		}
	}

	return &apistructs.PagingProjectDTO{Total: total, List: projectDTOs}, nil
}

// ListPublicProjects 获取公开项目列表
func (p *Project) ListPublicProjects(userID string, params *apistructs.ProjectListRequest) (
	*apistructs.PagingProjectDTO, error) {
	// 查找有权限的列表
	members, err := p.db.GetMembersByParentID(apistructs.ProjectScope, int64(params.OrgID), userID)
	if err != nil {
		return nil, errors.Errorf("failed to get permission when get projects, (%v)", err)
	}

	projectIDs := make([]uint64, 0, len(members))
	isManager := map[uint64]bool{} // 是否有管理权限
	isJoined := map[uint64]bool{}  // 是否加入了项目
	for i := range members {
		if members[i].ResourceKey == apistructs.RoleResourceKey {
			isJoined[uint64(members[i].ScopeID)] = true
			if members[i].ResourceValue == types.RoleProjectOwner ||
				members[i].ResourceValue == types.RoleProjectLead ||
				members[i].ResourceValue == types.RoleProjectPM {
				isManager[uint64(members[i].ScopeID)] = true
			} else {
				isManager[uint64(members[i].ScopeID)] = false
			}
		}
	}

	approves, err := p.db.ListUnblockApplicationApprove(params.OrgID)
	if err != nil {
		return nil, errors.Errorf("failed to ListUnblockApplicationApprove: %v", err)
	}

	org, err := p.db.GetOrg(int64(params.OrgID))
	if err != nil {
		return nil, errors.Errorf("failed to getorg(%d): %v", params.OrgID, err)
	}

	isOrgBlocked := false
	projectBlockStatus := map[uint64]string{}
	if org.BlockoutConfig.BlockDEV ||
		org.BlockoutConfig.BlockTEST ||
		org.BlockoutConfig.BlockStage ||
		org.BlockoutConfig.BlockProd {
		isOrgBlocked = true
		for _, i := range projectIDs {
			projectBlockStatus[i] = "blocked"
		}
	}

	for _, approve := range approves {
		if approve.Status == string(apistructs.ApprovalStatusPending) {
			projectBlockStatus[approve.TargetID] = "unblocking"
		}
	}
	// 获取项目列表
	total, projects, err := p.db.GetProjectsByOrgIDAndName(int64(params.OrgID), params)
	if err != nil {
		return nil, errors.Errorf("failed to get projects, (%v)", err)
	}

	// 转换成所需格式
	projectDTOs := make([]apistructs.ProjectDTO, 0, len(projects))
	projectIDs = make([]uint64, 0, len(projects))
	for i := range projects {
		projectDTOs = append(projectDTOs, p.convertToProjectDTO(isJoined[uint64(projects[i].ID)], &projects[i]))
		projectIDs = append(projectIDs, uint64(projects[i].ID))
	}
	resp, err := p.bdl.ProjectResource(projectIDs)
	if err != nil {
		return nil, err
	}

	// 获取每个项目的app信息
	now := time.Now()
	apps, err := p.db.GetApplicationsByProjectIDs(projectIDs)
	if err != nil {
		return nil, err
	}
	for _, app := range apps {
		if app.UnblockStart != nil && app.UnblockEnd != nil &&
			now.Before(*app.UnblockEnd) && now.After(*app.UnblockStart) {
			projectBlockStatus[uint64(app.ProjectID)] = "unblocked"
			break
		}
	}

	projectOwnerMap := make(map[uint64][]string)
	owners, err := p.db.GetMemberByScopeAndRole(apistructs.ProjectScope, projectIDs, []string{"owner"})
	if err != nil {
		return nil, err
	}
	for _, v := range owners {
		projectID := uint64(v.ScopeID)
		projectOwnerMap[projectID] = append(projectOwnerMap[projectID], v.UserID)
	}

	for i := range projectDTOs {
		if v, ok := resp.Data[projectDTOs[i].ID]; ok {
			projectDTOs[i].CpuServiceUsed = v.CpuServiceUsed
			projectDTOs[i].MemServiceUsed = v.MemServiceUsed
			projectDTOs[i].CpuAddonUsed = v.CpuAddonUsed
			projectDTOs[i].MemAddonUsed = v.MemAddonUsed
		}
		if v, ok := projectOwnerMap[projectDTOs[i].ID]; ok {
			projectDTOs[i].Owners = v
		}
	}

	for i := range projectDTOs {
		if isOrgBlocked {
			projectDTOs[i].BlockStatus = projectBlockStatus[projectDTOs[i].ID]
			canunblock := isManager[projectDTOs[i].ID]
			projectDTOs[i].CanUnblock = &canunblock
		}
		if isManager[projectDTOs[i].ID] {
			projectDTOs[i].CanManage = isManager[projectDTOs[i].ID]
		}
	}

	return &apistructs.PagingProjectDTO{Total: total, List: projectDTOs}, nil
}

// ListJoinedProjects 返回用户有权限的项目
func (p *Project) ListJoinedProjects(orgID int64, userID string, params *apistructs.ProjectListRequest) (
	*apistructs.PagingProjectDTO, error) {
	// 查找有权限的列表
	members, err := p.db.GetMembersByParentID(apistructs.ProjectScope, int64(params.OrgID), userID)
	if err != nil {
		return nil, errors.Errorf("failed to get permission when get projects, (%v)", err)
	}

	projectIDs := make([]uint64, 0, len(members))
	isManager := map[uint64]bool{}
	for i := range members {
		if members[i].ResourceKey == apistructs.RoleResourceKey {
			projectIDs = append(projectIDs, uint64(members[i].ScopeID))
			if members[i].ResourceValue == types.RoleProjectOwner ||
				members[i].ResourceValue == types.RoleProjectLead ||
				members[i].ResourceValue == types.RoleProjectPM {
				isManager[uint64(members[i].ScopeID)] = true
			} else {
				isManager[uint64(members[i].ScopeID)] = false
			}
		}
	}

	approves, err := p.db.ListUnblockApplicationApprove(uint64(orgID))
	if err != nil {
		return nil, errors.Errorf("failed to ListUnblockApplicationApprove: %v", err)
	}

	org, err := p.db.GetOrg(orgID)
	if err != nil {
		return nil, errors.Errorf("failed to getorg(%d): %v", orgID, err)
	}

	isOrgBlocked := false
	projectBlockStatus := map[uint64]string{}
	if org.BlockoutConfig.BlockDEV ||
		org.BlockoutConfig.BlockTEST ||
		org.BlockoutConfig.BlockStage ||
		org.BlockoutConfig.BlockProd {
		isOrgBlocked = true
		for _, i := range projectIDs {
			projectBlockStatus[i] = "blocked"
		}
	}

	for _, approve := range approves {
		if approve.Status == string(apistructs.ApprovalStatusPending) {
			projectBlockStatus[approve.TargetID] = "unblocking"
		}
	}
	// 获取项目列表
	total, projects, err := p.db.GetProjectsByIDs(projectIDs, params)
	if err != nil {
		return nil, errors.Errorf("failed to get projects, (%v)", err)
	}
	now := time.Now()
	for _, proj := range projects {
		apps, err := p.db.GetProjectApplications(proj.ID)
		if err != nil {
			return nil, errors.Errorf("failed to get app, proj(%d): %v", proj.ID, err)
		}
		for _, app := range apps {
			if app.UnblockStart != nil && app.UnblockEnd != nil &&
				now.Before(*app.UnblockEnd) && now.After(*app.UnblockStart) {
				projectBlockStatus[uint64(proj.ID)] = "unblocked"
				break
			}
		}
	}

	// 转换成所需格式
	projectDTOs := make([]apistructs.ProjectDTO, 0, len(projects))
	projectIDs = make([]uint64, 0, len(projects))
	for i := range projects {
		projectDTOs = append(projectDTOs, p.convertToProjectDTO(params.Joined, &projects[i]))
		projectIDs = append(projectIDs, uint64(projects[i].ID))
	}
	resp, err := p.bdl.ProjectResource(projectIDs)
	if err != nil {
		return nil, err
	}

	projectOwnerMap := make(map[uint64][]string)
	owners, err := p.db.GetMemberByScopeAndRole(apistructs.ProjectScope, projectIDs, []string{"owner"})
	if err != nil {
		return nil, err
	}
	for _, v := range owners {
		projectID := uint64(v.ScopeID)
		projectOwnerMap[projectID] = append(projectOwnerMap[projectID], v.UserID)
	}

	for i := range projectDTOs {
		if v, ok := resp.Data[projectDTOs[i].ID]; ok {
			projectDTOs[i].CpuServiceUsed = v.CpuServiceUsed
			projectDTOs[i].MemServiceUsed = v.MemServiceUsed
			projectDTOs[i].CpuAddonUsed = v.CpuAddonUsed
			projectDTOs[i].MemAddonUsed = v.MemAddonUsed
		}
		if v, ok := projectOwnerMap[projectDTOs[i].ID]; ok {
			projectDTOs[i].Owners = v
		}
	}

	for i := range projectDTOs {
		if isOrgBlocked {
			projectDTOs[i].BlockStatus = projectBlockStatus[projectDTOs[i].ID]
			canunblock := isManager[projectDTOs[i].ID]
			projectDTOs[i].CanUnblock = &canunblock
		}
		if isManager[projectDTOs[i].ID] {
			projectDTOs[i].CanManage = isManager[projectDTOs[i].ID]
		}
	}

	return &apistructs.PagingProjectDTO{Total: total, List: projectDTOs}, nil
}

// ReferCluster 检查 cluster 是否被某个项目所使用
func (p *Project) ReferCluster(clusterName string) bool {
	projects, err := p.db.ListProjectByCluster(clusterName)
	if err != nil {
		logrus.Warnf("check cluster if referred by project")
		return true
	}
	if len(projects) > 0 {
		return true
	}

	return false
}

// UpdateProjectFunction 更新项目的功能开关
func (p *Project) UpdateProjectFunction(req *apistructs.ProjectFunctionSetRequest) (int64, error) {
	// 检查待更新的project是否存在
	project, err := p.db.GetProjectByID(int64(req.ProjectID))
	if err != nil {
		return 0, errors.Wrap(err, "failed to update project function")
	}

	var pf map[apistructs.ProjectFunction]bool
	json.Unmarshal([]byte(project.Functions), &pf)
	for k, v := range req.ProjectFunction {
		pf[k] = v
	}
	newFunction, err := json.Marshal(pf)
	if err != nil {
		return 0, errors.Wrap(err, "failed to update project function")
	}
	project.Functions = string(newFunction)

	if err = p.db.UpdateProject(&project); err != nil {
		logrus.Warnf("failed to update project, (%v)", err)
		return 0, errors.Errorf("failed to update project function")
	}

	return project.ID, nil
}

// UpdateProjectActiveTime 更新项目活跃时间
func (p *Project) UpdateProjectActiveTime(req *apistructs.ProjectActiveTimeUpdateRequest) error {
	// 检查待更新的project是否存在
	project, err := p.db.GetProjectByID(int64(req.ProjectID))
	if err != nil {
		return errors.Wrap(err, "failed to update project function")
	}

	if project.ActiveTime.Unix() > req.ActiveTime.Unix() {
		return nil
	}

	project.ActiveTime = req.ActiveTime
	if err = p.db.UpdateProject(&project); err != nil {
		logrus.Warnf("failed to update project, (%v)", err)
		return errors.Errorf("failed to update project function")
	}

	return nil
}

func checkRollbackConfig(rollbackConfig *map[string]int) error {
	// DEV/TEST/STAGING/PROD
	l := len(*rollbackConfig)

	// if empty then don't update
	if l == 0 {
		return nil
	}

	// check
	if l != 4 {
		return errors.Errorf("invalid param(clusterConfig is empty)")
	}

	for key := range *rollbackConfig {
		switch key {
		case string(types.DevWorkspace), string(types.TestWorkspace), string(types.StagingWorkspace),
			string(types.ProdWorkspace):
		default:
			return errors.Errorf("invalid param, rollback config: %s", key)

		}
	}
	return nil
}

// initRollbackConfig init rollback config when create a project
func initRollbackConfig(rollbackConfig *map[string]int) error {
	if len(*rollbackConfig) != 4 {
		*rollbackConfig = map[string]int{
			string(types.DevWorkspace):     5,
			string(types.TestWorkspace):    5,
			string(types.StagingWorkspace): 5,
			string(types.ProdWorkspace):    5,
		}
	}
	return checkRollbackConfig(rollbackConfig)
}

func (p *Project) convertToProjectDTO(joined bool, project *model.Project) apistructs.ProjectDTO {
	var rollbackConfig map[string]int
	if err := json.Unmarshal([]byte(project.RollbackConfig), &rollbackConfig); err != nil {
		rollbackConfig = make(map[string]int)
	}

	total, _ := p.db.GetApplicationCountByProjectID(project.ID)

	projectDto := apistructs.ProjectDTO{
		ID:          uint64(project.ID),
		Name:        project.Name,
		DisplayName: project.DisplayName,
		Desc:        project.Desc,
		Logo:        filehelper.APIFileUrlRetriever(project.Logo),
		OrgID:       uint64(project.OrgID),
		Joined:      joined,
		Creator:     project.UserID,
		DDHook:      project.DDHook,
		Stats: apistructs.ProjectStats{
			CountApplications: int(total),
		},
		ClusterConfig:  nil,
		RollbackConfig: rollbackConfig,
		CpuQuota:       project.CpuQuota,
		MemQuota:       project.MemQuota,
		CreatedAt:      project.CreatedAt,
		UpdatedAt:      project.UpdatedAt,
		ActiveTime:     project.ActiveTime.Format("2006-01-02 15:04:05"),
		Owners:         []string{},
		IsPublic:       project.IsPublic,
		Type:           project.Type,
	}
	if projectDto.DisplayName == "" {
		projectDto.DisplayName = projectDto.Name
	}

	return projectDto
}

// GetProjectStats 获取项目状态
func (p *Project) GetProjectStats(projectID int64) (*apistructs.ProjectStats, error) {
	totalApp, err := p.db.GetApplicationCountByProjectID(projectID)
	if err != nil {
		return nil, errors.Errorf("get project states err: get app err: %v", err)
	}
	totalMembers, _, err := p.db.GetMembersWithoutExtraByScope(apistructs.ProjectScope, projectID)
	if err != nil {
		return nil, errors.Errorf("get project states err: get member err: %v", err)
	}
	return &apistructs.ProjectStats{
		CountApplications:      int(totalApp),
		CountMembers:           totalMembers,
		TotalApplicationsCount: int(totalApp),
		TotalMembersCount:      totalMembers,
	}, nil
}

// GetProjectNSInfo 获取项目级别命名空间信息
func (p *Project) GetProjectNSInfo(projectID int64) (*apistructs.ProjectNameSpaceInfo, error) {
	prj, err := p.db.GetProjectByID(projectID)
	if err != nil {
		return nil, err
	}
	prjIDStr := strconv.FormatInt(projectID, 10)

	prjNsInfo := &apistructs.ProjectNameSpaceInfo{
		Enabled:    false,
		Namespaces: make(map[string]string, 0),
	}

	if prj.EnableNS {
		prjNsInfo.Enabled = true
		prjNsInfo.Namespaces = genProjectNamespace(prjIDStr)
	}

	return prjNsInfo, nil
}

// help func
// genProjectNamespace 生成项目级命名空间
func genProjectNamespace(prjIDStr string) map[string]string {
	return map[string]string{"DEV": "project-" + prjIDStr + "-dev", "TEST": "project-" + prjIDStr + "-test",
		"STAGING": "project-" + prjIDStr + "-staging", "PROD": "project-" + prjIDStr + "-prod"}
}

func (p *Project) GetMyProjectIDList(parentID int64, userID string) ([]uint64, error) {
	members, err := p.db.GetMembersByParentID(apistructs.ProjectScope, parentID, userID)
	if err != nil {
		return nil, errors.Errorf("failed to get permission when get projects, (%v)", err)
	}

	projectIDList := make([]uint64, 0, len(members))
	for i := range members {
		if members[i].ResourceKey == apistructs.RoleResourceKey {
			projectIDList = append(projectIDList, uint64(members[i].ScopeID))
		}
	}
	return projectIDList, nil
}

func (p *Project) GetProjectIDListByStates(req apistructs.IssuePagingRequest, projectIDList []uint64) (int, []apistructs.ProjectDTO, error) {
	var res []apistructs.ProjectDTO
	total, pros, err := p.db.GetProjectIDListByStates(req, projectIDList)
	if err != nil {
		return total, res, err
	}
	for _, v := range pros {
		proDTO := p.convertToProjectDTO(true, &v)
		res = append(res, proDTO)
	}
	return total, res, nil
}

func (p *Project) GetQuotaOnClusters(orgID int64, clusterNames []string) (*apistructs.GetQuotaOnClustersResponse, error) {
	l := logrus.WithField("func", "GetQuotaOnClusters")

	var response = new(apistructs.GetQuotaOnClustersResponse)
	response.ClusterNames = clusterNames

	if len(clusterNames) == 0 {
		l.Warnln("no clusters for GetQuotaOnClusters")
		return response, nil
	}

	var clusterNamesM = make(map[string]bool)
	for _, clusterName := range clusterNames {
		clusterNamesM[clusterName] = true
	}

	// query all projects
	var projects []*model.Project
	if err := p.db.Find(&projects, map[string]interface{}{"org_id": orgID}).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			logrus.WithError(err).Warnln("project record not found")
			return response, nil
		}
		err = errors.Wrap(err, "failed to Find projects")
		l.WithError(err).Errorln()
		return nil, err
	}

	// query all project quota
	var projectIDs []int64
	for _, project := range projects {
		projectIDs = append(projectIDs, project.ID)
	}
	var projectsQuota []*model.ProjectQuota
	if err := p.db.Where("project_id IN (?)", projectIDs).Find(&projectsQuota).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			l.WithError(err).Warnln("quota record not found")
			return response, nil
		}
		err = errors.Wrap(err, "failed to Find project quota")
		return nil, err
	}
	var projectsQuotaConfigs = make(map[int64]*model.ProjectQuota)
	for _, projectQuota := range projectsQuota {
		projectsQuotaConfigs[int64(projectQuota.ProjectID)] = projectQuota
	}

	var ownerM = make(map[string]*apistructs.OwnerQuotaOnClusters)
	for _, project := range projects {
		// query project owner
		var member = &model.Member{
			UserID: "0",
			Name:   "unknown",
			Nick:   "unknown",
		}
		memberListReq := apistructs.MemberListRequest{
			ScopeType: "project",
			ScopeID:   project.ID,
			Roles:     []string{"Owner", "Lead"},
			Labels:    nil,
			Q:         "",
			PageNo:    1,
			PageSize:  1,
		}
		switch _, members, err := p.db.GetMembersByParam(&memberListReq); {
		case err != nil:
			l.WithError(err).WithField("memberListReq", memberListReq).Warnln("failed to GetMembersByParam")
		case len(members) == 0:
			l.WithError(err).WithField("memberListReq", memberListReq).Warnln("not found owner for the project")
		default:
			mb, ok := getMemberFromMembers(members, "Owner")
			if ok {
				member = mb
				break
			}
			if mb, ok = getMemberFromMembers(members, "Lead"); ok {
				member = mb
			}
		}

		owner, ok := ownerM[member.UserID]
		if !ok {
			userID, err := strconv.ParseInt(member.UserID, 10, 64)
			if err != nil {
				err = errors.Wrap(err, "the format of owner userID is not valid")
			}
			owner = &apistructs.OwnerQuotaOnClusters{
				ID:       uint64(userID),
				Name:     member.Name,
				Nickname: member.Nick,
				CPUQuota: 0,
				MemQuota: 0,
				Projects: nil,
			}
			ownerM[member.UserID] = owner
		}

		projectQuotaOnCluster := apistructs.ProjectQuotaOnClusters{
			ID:          uint64(project.ID),
			Name:        project.Name,
			DisplayName: project.DisplayName,
			CPUQuota:    0,
			MemQuota:    0,
		}
		// if the project has quota config, accumulate it
		if config, ok := projectsQuotaConfigs[project.ID]; ok {
			// if a cluster of the project is in the specified clusters, its cpu and mem will be accumulated
			for clusterName, quota := range map[string][2]uint64{
				config.ProdClusterName:    {config.ProdCPUQuota, config.ProdMemQuota},
				config.StagingClusterName: {config.StagingCPUQuota, config.StagingMemQuota},
				config.TestClusterName:    {config.TestCPUQuota, config.TestMemQuota},
				config.DevClusterName:     {config.DevCPUQuota, config.DevMemQuota},
			} {
				if _, ok := clusterNamesM[clusterName]; ok {
					projectQuotaOnCluster.AccuQuota(quota[0], quota[1])
				}
			}
		}
		owner.Projects = append(owner.Projects, &projectQuotaOnCluster)
	}

	for _, owner := range ownerM {
		response.Owners = append(response.Owners, owner)
	}

	response.ReCalcu()

	return response, nil
}

func (p *Project) GetNamespacesBelongsTo(ctx context.Context) (*apistructs.GetProjectsNamesapcesResponseData, error) {
	l := logrus.WithField("func", "GetNamespacesBelongsTo")

	langCodes, _ := ctx.Value("lang_codes").(i18n.LanguageCodes)
	unknownName := p.trans.Text(langCodes, "OwnerUnknown")

	// 1）查找 s_pod_info
	var projectsM = make(map[uint64]map[string][]string)
	var podInfos []*apistructs.PodInfo
	if err := p.db.Debug().Find(&podInfos).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			err = errors.Wrap(err, "failed to Find podInfos")
			l.WithError(err).Errorln()
			return nil, err
		}
	}

	for _, podInfo := range podInfos {
		projectID, err := strconv.ParseUint(podInfo.ProjectID, 10, 64)
		if err != nil {
			continue
		}
		clusters, ok := projectsM[projectID]
		if !ok {
			clusters = make(map[string][]string)
		}
		clusters[podInfo.Cluster] = append(clusters[podInfo.Cluster], podInfo.K8sNamespace)
		projectsM[projectID] = clusters
	}

	// 2) 查找 project_namespace
	var projectNamespaces []*apistructs.ProjectNamespaceModel
	if err := p.db.Find(&projectNamespaces).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			err = errors.Wrap(err, "failed to Find projectNamespace")
			l.WithError(err).Errorln()
			return nil, err
		}
	}
	for _, projectNamespace := range projectNamespaces {
		clusters, ok := projectsM[projectNamespace.ProjectID]
		if !ok {
			clusters = make(map[string][]string)
		}
		clusters[projectNamespace.ClusterName] = append(clusters[projectNamespace.ClusterName], projectNamespace.K8sNamespace)
		projectsM[projectNamespace.ProjectID] = clusters
	}

	var data apistructs.GetProjectsNamesapcesResponseData

	// 3) 查询 quota
	for projectID, clusterNamespaces := range projectsM {
		var project model.Project
		if err := p.db.First(&project, map[string]interface{}{"id": projectID}).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				l.WithError(err).WithField("id", projectID).Warnln("failed to First project")
				continue
			}
		}

		// query project owner
		var member = &model.Member{
			UserID: "0",
			Name:   unknownName,
			Nick:   unknownName,
		}
		memberListReq := apistructs.MemberListRequest{
			ScopeType: "project",
			ScopeID:   project.ID,
			Roles:     []string{"Owner", "Lead"},
			Labels:    nil,
			Q:         "",
			PageNo:    1,
			PageSize:  1,
		}
		switch _, members, err := p.db.GetMembersByParam(&memberListReq); {
		case err != nil:
			l.WithError(err).WithField("memberListReq", memberListReq).Warnln("failed to GetMembersByParam")
		case len(members) == 0:
			l.WithError(err).WithField("memberListReq", memberListReq).Warnln("not found owner for the project")
		default:
			mb, ok := getMemberFromMembers(members, "Owner")
			if ok {
				member = mb
				break
			}
			if mb, ok = getMemberFromMembers(members, "Lead"); ok {
				member = mb
			}
		}
		userID, err := strconv.ParseUint(member.UserID, 10, 64)
		if err != nil {
			l.Warnln("failed to parse member.UserID")
			continue
		}

		// query quota
		var quota apistructs.ProjectQuota
		if err := p.db.First(&quota, map[string]interface{}{"project_id": projectID}).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				err = errors.Wrap(err, "failed to First project quota")
				l.WithError(err).WithField("project_id", projectID).Errorln()
				return nil, err
			}
		}
		var item = apistructs.ProjectNamespaces{
			ProjectID:          uint(project.ID),
			ProjectName:        project.Name,
			ProjectDisplayName: project.DisplayName,
			ProjectDesc:        project.Desc,
			OwnerUserID:        uint(userID),
			OwnerUserName:      member.Name,
			OwnerUserNickname:  member.Nick,
			CPUQuota:           uint64(quota.ProdCPUQuota + quota.StagingCPUQuota + quota.TestCPUQuota + quota.DevCPUQuota),
			MemQuota:           uint64(quota.ProdMemQuota + quota.StagingMemQuota + quota.TestMemQuota + quota.DevMemQuota),
			Clusters:           clusterNamespaces,
		}
		data.List = append(data.List, &item)
	}
	data.Total = uint32(len(data.List))

	return &data, nil
}

func getMemberFromMembers(members []model.Member, role string) (*model.Member, bool) {
	for _, member := range members {
		for _, role_ := range member.Roles {
			if strings.EqualFold(role, role_) {
				return &member, true
			}
		}
	}
	return nil, false
}
