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
	"math"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/core-services/conf"
	"github.com/erda-project/erda/modules/core-services/dao"
	"github.com/erda-project/erda/modules/core-services/model"
	"github.com/erda-project/erda/modules/core-services/types"
	"github.com/erda-project/erda/pkg/crypto/uuid"
	"github.com/erda-project/erda/pkg/filehelper"
	"github.com/erda-project/erda/pkg/numeral"
	"github.com/erda-project/erda/pkg/ucauth"
)

// Project 资源对象操作封装
type Project struct {
	db  *dao.DBClient
	uc  *ucauth.UCClient
	bdl *bundle.Bundle
}

// Option 定义 Project 对象的配置选项
type Option func(*Project)

// New 新建 Project 实例，通过 Project 实例操作企业资源
func New(options ...Option) *Project {
	p := &Project{}
	for _, op := range options {
		op(p)
	}
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

	if err := checkClusterConfig(createReq.ClusterConfig); err != nil {
		return nil, err
	}
	clusterConfig, err := json.Marshal(createReq.ClusterConfig)
	if err != nil {
		logrus.Infof("failed to marshal clusterConfig, (%v)", err)
		return nil, errors.Errorf("failed to marshal clusterConfig")
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
		Type:           string(createReq.Template),
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

func (p *Project) GetModelProjectsMap(projectIDs []uint64) (map[int64]*model.Project, error) {
	_, projects, err := p.db.GetProjectsByIDs(projectIDs, &apistructs.ProjectListRequest{
		PageNo:   1,
		PageSize: len(projectIDs),
	})
	if err != nil {
		return nil, errors.Errorf("failed to get projects, (%v)", err)
	}

	projectMap := make(map[int64]*model.Project)
	for _, p := range projects {
		projectMap[p.ID] = &p
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
		Logo:        filehelper.APIFileUrlRetriever(project.Logo),
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
