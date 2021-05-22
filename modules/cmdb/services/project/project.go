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

// Package project 封装项目资源相关操作
package project

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmdb/conf"
	"github.com/erda-project/erda/modules/cmdb/dao"
	"github.com/erda-project/erda/modules/cmdb/model"
	"github.com/erda-project/erda/modules/cmdb/types"
	"github.com/erda-project/erda/pkg/cron"
	"github.com/erda-project/erda/pkg/numeral"
	"github.com/erda-project/erda/pkg/ucauth"
	"github.com/erda-project/erda/pkg/uuid"
)

// Project 资源对象操作封装
type Project struct {
	db                *dao.DBClient
	uc                *ucauth.UCClient
	bdl               *bundle.Bundle
	ProjectStatsCache *sync.Map
}

// Option 定义 Project 对象的配置选项
type Option func(*Project)

// New 新建 Project 实例，通过 Project 实例操作企业资源
func New(options ...Option) *Project {
	p := &Project{
		ProjectStatsCache: &sync.Map{},
	}
	for _, op := range options {
		op(p)
	}

	p.setProjectStatsCache()
	return p
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(p *Project) {
		p.db = db
	}
}

// WithUCClient 配置 uc client
func WithUCClient(uc *ucauth.UCClient) Option {
	return func(p *Project) {
		p.uc = uc
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(p *Project) {
		p.bdl = bdl
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
		Sender:  bundle.SenderCMDB,
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

	if err := checkClusterConfig(createReq.ClusterConfig); err != nil {
		return nil, err
	}
	clusterConfig, err := json.Marshal(createReq.ClusterConfig)
	if err != nil {
		logrus.Infof("failed to marshal clusterConfig, (%v)", err)
		return nil, errors.Errorf("failed to marshal clusterConfig")
	}
	if err := checkRollbackConfig(&createReq.RollbackConfig); err != nil {
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

	// 添加项目至DB
	project = &model.Project{
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
		ActiveTime:     time.Now(),
		EnableNS:       conf.EnableNS(),
	}
	if err = p.db.CreateProject(project); err != nil {
		logrus.Warnf("failed to insert project to db, (%v)", err)
		return nil, errors.Errorf("failed to insert project to db")
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
		if err = p.db.CreateMember(&member); err != nil {
			logrus.Warnf("failed to add member to db when create project, (%v)", err)
		}
		if err = p.db.CreateMemberExtra(&memberExtra); err != nil {
			logrus.Warnf("failed to add member roles to db when create project, (%v)", err)
		}
	}
	if err := p.InitProjectState(project.ID); err != nil {
		logrus.Warnf("failed to add state to db when create project, (%v)", err)
	}
	return project, nil
}

// UpdateWithEvent 更新项目 & 发送事件
func (p *Project) UpdateWithEvent(projectID int64, updateReq *apistructs.ProjectUpdateBody) error {
	// 更新项目
	project, err := p.Update(projectID, updateReq)
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
		Sender:  bundle.SenderCMDB,
		Content: *project,
	}
	// 发送项目更新事件
	if err = p.bdl.CreateEvent(ev); err != nil {
		logrus.Warnf("failed to send project update event, (%v)", err)
	}

	return nil
}

// Update 更新项目
func (p *Project) Update(projectID int64, updateReq *apistructs.ProjectUpdateBody) (*model.Project, error) {
	if err := checkClusterConfig(updateReq.ClusterConfig); err != nil {
		return nil, err
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

	if err = p.db.UpdateProject(&project); err != nil {
		logrus.Warnf("failed to update project, (%v)", err)
		return nil, errors.Errorf("failed to update project")
	}

	return &project, nil
}

func patchProject(project *model.Project, updateReq *apistructs.ProjectUpdateBody) error {
	clusterConfig, err := json.Marshal(updateReq.ClusterConfig)
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

	if len(updateReq.ClusterConfig) != 0 {
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
		Sender:  bundle.SenderCMDB,
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
	// 检查项目下是否有应用，无应用时可删除
	if count, err := p.db.GetApplicationCountByProjectID(projectID); err != nil || count > 0 {
		return nil, errors.Errorf("failed to delete project(there exists applications)")
	}

	project, err := p.db.GetProjectByID(projectID)
	if err != nil {
		return nil, errors.Errorf("failed to get project, (%v)", err)
	}

	if err = p.db.DeleteProject(projectID); err != nil {
		return nil, errors.Errorf("failed to delete project, (%v)", err)
	}
	logrus.Infof("deleted project %d", projectID)

	// 删除权限表记录
	if err = p.db.DeleteMembersByScope(apistructs.ProjectScope, projectID); err != nil {
		logrus.Warnf("failed to delete members, (%v)", err)
	}
	if err = p.db.DeleteMemberExtraByScope(apistructs.ProjectScope, projectID); err != nil {
		logrus.Warnf("failed to delete members extra, (%v)", err)
	}

	if err = p.db.DeleteBranchRuleByScope(apistructs.ProjectScope, projectID); err != nil {
		logrus.Warnf("failed to delete project branch rules, (%v)", err)
	}

	// 删除状态表
	if err = p.db.DeleteIssuesStateByProjectID(projectID); err != nil {
		logrus.Warnf("failed to delete project state, (%v)", err)
	}
	return &project, nil
}

// Get 获取项目
func (p *Project) Get(projectID int64) (*apistructs.ProjectDTO, error) {
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
	return &projectDTO, nil
}

// GetModelProject 获取项目
func (p *Project) GetModelProject(projectID int64) (*model.Project, error) {
	project, err := p.db.GetProjectByID(projectID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	return &project, nil
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

// ListProjects
func (p *Project) GetAllProjects() ([]model.Project, error) {
	return p.db.GetAllProjects()
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

// 检查cluster config合法性
func checkClusterConfig(clusterConfig map[string]string) error {
	// DEV/TEST/STAGING/PROD四个环境集群配置
	l := len(clusterConfig)
	// 空则不配置
	if l == 0 {
		return nil
	}

	// check
	if l != 4 {
		return errors.Errorf("invalid param(clusterConfig is empty)")
	}
	for key := range clusterConfig {
		switch key {
		case string(types.DevWorkspace), string(types.TestWorkspace), string(types.StagingWorkspace),
			string(types.ProdWorkspace):
		default:
			return errors.Errorf("invalid param, cluster config: %s", key)
		}
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
func (p *Project) convertToProjectDTO(joined bool, project *model.Project) apistructs.ProjectDTO {
	var clusterConfig map[string]string
	if err := json.Unmarshal([]byte(project.ClusterConfig), &clusterConfig); err != nil {
		clusterConfig = make(map[string]string)
	}
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
		Logo:        project.Logo,
		OrgID:       uint64(project.OrgID),
		Joined:      joined,
		Creator:     project.UserID,
		DDHook:      project.DDHook,
		Stats: apistructs.ProjectStats{
			CountApplications: int(total),
		},
		ClusterConfig:  clusterConfig,
		RollbackConfig: rollbackConfig,
		CpuQuota:       project.CpuQuota,
		MemQuota:       project.MemQuota,
		CreatedAt:      project.CreatedAt,
		UpdatedAt:      project.UpdatedAt,
		ActiveTime:     project.ActiveTime.Format("2006-01-02 15:04:05"),
		Owners:         []string{},
		IsPublic:       project.IsPublic,
	}
	if projectDto.DisplayName == "" {
		projectDto.DisplayName = projectDto.Name
	}

	return projectDto
}

// setProjectStatsCache 设置项目状态缓存
func (p *Project) setProjectStatsCache() {
	c := cron.New()
	if err := c.AddFunc(conf.ProjectStatsCacheCron(), func() {
		// 清空缓存
		logrus.Info("start set project stats")
		p.ProjectStatsCache = new(sync.Map)
		// prjs, err := p.db.GetAllProjects()
		// if err != nil {
		// 	logrus.Errorf("get project stats err: get all project err: %v", err)
		// }
		//
		// for _, prj := range prjs {
		// 	stats, err := p.GetProjectStats(prj.ID)
		// 	if err != nil {
		// 		logrus.Errorf("get project %v stats err: %v", prj.ID, err)
		// 	}
		// 	p.ProjectStatsCache.Store(prj.ID, stats)
		// }
	}); err != nil {
		logrus.Errorf("cron set setProjectStatsCache failed: %v", err)
	}

	c.Start()
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
	iterations, err := p.db.FindIterations(uint64(projectID))
	if err != nil {
		return nil, errors.Errorf("get project states err: get iterations err: %v", err)
	}
	totalIterations := len(iterations)

	runningIterations, planningIterations := make([]int64, 0), make(map[int64]bool, 0)
	now := time.Now()
	for i := 0; i < totalIterations; i++ {
		if !iterations[i].StartedAt.After(now) && iterations[i].FinishedAt.After(now) {
			runningIterations = append(runningIterations, iterations[i].ID)
		}

		if iterations[i].StartedAt.After(now) {
			planningIterations[iterations[i].ID] = true
		}
	}

	var totalManHour, usedManHour, planningManHour, totalBug, doneBug int64
	totalIssues, _, err := p.db.PagingIssues(apistructs.IssuePagingRequest{
		IssueListRequest: apistructs.IssueListRequest{
			ProjectID: uint64(projectID),
			Type:      []apistructs.IssueType{apistructs.IssueTypeBug, apistructs.IssueTypeTask},
			External:  true,
		},
		PageNo:   1,
		PageSize: 99999,
	}, false)
	if err != nil {
		return nil, errors.Errorf("get project states err: get issues err: %v", err)
	}

	// 事件状态map
	closedBugStatsMap := make(map[int64]struct{}, 0)
	bugState, err := p.db.GetClosedBugState(projectID)
	if err != nil {
		return nil, errors.Errorf("get project states err: get issues stats err: %v", err)
	}
	for _, v := range bugState {
		closedBugStatsMap[v.ID] = struct{}{}
	}

	for _, v := range totalIssues {
		var manHour apistructs.IssueManHour
		json.Unmarshal([]byte(v.ManHour), &manHour)
		// set total and used man-hour
		totalManHour += manHour.EstimateTime
		usedManHour += manHour.ElapsedTime
		// set plannning man-hour
		if _, ok := planningIterations[v.IterationID]; ok {
			planningManHour += manHour.EstimateTime
		}
		if v.Type == apistructs.IssueTypeBug {
			if _, ok := closedBugStatsMap[v.State]; ok {
				doneBug++
			}
			totalBug++
		}
	}

	tManHour, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(totalManHour)/480), 64)
	uManHour, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(usedManHour)/480), 64)
	pManHour, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(planningManHour)/480), 64)
	var dBugPer float64 = 100
	if totalBug != 0 {
		dBugPer, _ = strconv.ParseFloat(fmt.Sprintf("%.0f", float64(doneBug)*100/float64(totalBug)), 64)
	}

	return &apistructs.ProjectStats{
		CountApplications:       int(totalApp),
		CountMembers:            totalMembers,
		TotalApplicationsCount:  int(totalApp),
		TotalMembersCount:       totalMembers,
		TotalIterationsCount:    totalIterations,
		RunningIterationsCount:  len(runningIterations),
		PlanningIterationsCount: len(planningIterations),
		TotalManHourCount:       tManHour,
		UsedManHourCount:        uManHour,
		PlanningManHourCount:    pManHour,
		DoneBugCount:            doneBug,
		TotalBugCount:           totalBug,
		DoneBugPercent:          dBugPer,
	}, nil
}

func (p *Project) InitProjectState(projectID int64) error {
	var (
		states    []dao.IssueState
		relations []dao.IssueStateRelation
	)
	relation := []int64{
		0, 1, 0, 2, 0, 3, 1, 2, 1, 3, 2, 3, 1, 0, 2, 0, 3, 0, 2, 1, 3, 1, 3, 2,
		4, 5, 5, 6,
		7, 8, 8, 9,
		10, 13, 14, 13, 11, 14, 12, 14, 13, 14, 10, 11, 14, 11, 10, 12, 14, 12, 11, 15, 12, 15, 13, 15,
		16, 19, 20, 19, 17, 20, 18, 20, 19, 20, 16, 17, 20, 17, 16, 18, 20, 18, 17, 21, 18, 21, 19, 21,
	}
	name := []string{
		"待处理", "进行中", "测试中", "已完成",
		"待处理", "进行中", "已完成",
		"待处理", "进行中", "已完成",
		"待处理", "无需修复", "重复提交", "已解决", "重新打开", "已关闭",
		"待处理", "无需修复", "重复提交", "已解决", "重新打开", "已关闭",
	}
	belong := []apistructs.IssueStateBelong{
		"OPEN", "WORKING", "WORKING", "DONE",
		"OPEN", "WORKING", "DONE",
		"OPEN", "WORKING", "DONE",
		"OPEN", "WONTFIX", "WONTFIX", "RESOLVED", "REOPEN", "CLOSED",
		"OPEN", "WONTFIX", "WONTFIX", "RESOLVED", "REOPEN", "CLOSED",
	}
	index := []int64{
		0, 1, 2, 3,
		0, 1, 2,
		0, 1, 2,
		0, 1, 2, 3, 4, 5,
		0, 1, 2, 3, 4, 5,
	}
	// state
	for i := 0; i < 22; i++ {
		states = append(states, dao.IssueState{
			ProjectID: uint64(projectID),
			Name:      name[i],
			Belong:    belong[i],
			Index:     index[i],
			Role:      "Ops,Dev,QA,Owner,Lead",
		})
		if i < 4 {
			states[i].IssueType = apistructs.IssueTypeRequirement
		} else if i < 7 {
			states[i].IssueType = apistructs.IssueTypeTask
		} else if i < 10 {
			states[i].IssueType = apistructs.IssueTypeEpic
		} else if i < 16 {
			states[i].IssueType = apistructs.IssueTypeBug
		} else if i < 22 {
			states[i].IssueType = apistructs.IssueTypeTicket
		}
		if err := p.db.CreateIssuesState(&states[i]); err != nil {
			return err
		}
	}
	// state relation
	for i := 0; i < 40; i++ {
		relations = append(relations, dao.IssueStateRelation{
			ProjectID:    projectID,
			StartStateID: states[relation[i*2]].ID,
			EndStateID:   states[relation[i*2+1]].ID,
		})
		if i < 12 {
			relations[i].IssueType = apistructs.IssueTypeRequirement
		} else if i < 14 {
			relations[i].IssueType = apistructs.IssueTypeTask
		} else if i < 16 {
			relations[i].IssueType = apistructs.IssueTypeEpic
		} else if i < 28 {
			relations[i].IssueType = apistructs.IssueTypeBug
		} else if i < 40 {
			relations[i].IssueType = apistructs.IssueTypeTicket
		}
	}
	return p.db.UpdateIssueStateRelations(projectID, apistructs.IssueTypeTask, relations)
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

func (p *Project) GetMyProjectIDList(scopeType apistructs.ScopeType, parentID int64, userID string) ([]uint64, error) {
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
