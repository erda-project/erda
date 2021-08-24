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

package dao

import (
	"fmt"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-proto-go/msp/tenant/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/model"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateProject 创建项目
func (client *DBClient) CreateProject(project *model.Project) error {
	return client.Create(project).Error
}

// UpdateProject 更新项目
func (client *DBClient) UpdateProject(project *model.Project) error {
	return client.Save(project).Error
}

// DeleteProject 删除项目
func (client *DBClient) DeleteProject(projectID int64) error {
	return client.Where("id = ?", projectID).Delete(&model.Project{}).Error
}

// GetProjectByID 根据projectID获取项目信息
func (client *DBClient) GetProjectByID(projectID int64) (model.Project, error) {
	var project model.Project
	if err := client.Where("id = ?", projectID).Find(&project).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return project, ErrNotFoundProject
		}
		return project, err
	}
	return project, nil
}

// GetProjectsByOrgIDAndName 根据orgID与名称获取项目列表
func (client *DBClient) GetProjectsByOrgIDAndName(orgID int64, params *apistructs.ProjectListRequest) (
	int, []model.Project, error) {
	var (
		projects []model.Project
		total    int
	)
	db := client.Where("org_id = ?", orgID)
	if params.IsPublic {
		db = db.Where("is_public = ?", params.IsPublic)
	}
	if params.Name != "" {
		db = db.Where("name = ?", params.Name)
	}
	if params.Query != "" {
		db = db.Where("(name LIKE ? OR display_name LIKE ?)", strutil.Concat("%", params.Query, "%"), strutil.Concat("%", params.Query, "%"))
	}
	if params.OrderBy != "" {
		if params.Asc {
			db = db.Order(fmt.Sprintf("%s", params.OrderBy))
		} else {
			db = db.Order(fmt.Sprintf("%s DESC", params.OrderBy))
		}
	} else {
		db = db.Order("active_time DESC")
	}

	if err := db.Offset((params.PageNo - 1) * params.PageSize).Limit(params.PageSize).
		Find(&projects).Offset(0).Limit(-1).Count(&total).Error; err != nil {
		return 0, nil, err
	}

	return total, projects, nil
}

// GetProjectsByIDs 根据projectIDs获取项目列表
func (client *DBClient) GetProjectsByIDs(projectIDs []uint64, params *apistructs.ProjectListRequest) (
	int, []model.Project, error) {
	var (
		total    int
		projects []model.Project
	)
	db := client.Where("id in (?)", projectIDs)
	if params.Name != "" {
		db = db.Where("name = ?", params.Name)
	}
	if params.Query != "" {
		db = db.Where("(name LIKE ? OR display_name LIKE ?)", strutil.Concat("%", params.Query, "%"), strutil.Concat("%", params.Query, "%"))
	}
	db = db.Where("`type` != ?", pb.Type_MSP.String())
	if params.OrderBy != "" {
		if params.Asc {
			db = db.Order(fmt.Sprintf("%s", params.OrderBy))
		} else {
			db = db.Order(fmt.Sprintf("%s DESC", params.OrderBy))
		}
	} else {
		db = db.Order("active_time DESC")
	}
	if err := db.Offset((params.PageNo - 1) * params.PageSize).Limit(params.PageSize).
		Find(&projects).Offset(0).Limit(-1).Count(&total).Error; err != nil {
		return 0, nil, err
	}

	return total, projects, nil
}

// GetProjectByOrgAndName 根据orgID & 项目名称 获取项目
func (client *DBClient) GetProjectByOrgAndName(orgID int64, name string) (*model.Project, error) {
	var project model.Project
	if err := client.Where("org_id = ?", orgID).
		Where("name = ?", name).Find(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// GetAllProjects get all projects
func (client *DBClient) GetAllProjects() ([]model.Project, error) {
	var projects []model.Project
	if err := client.Model(model.Project{}).Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

// ListProjectByOrgID 根据 orgID 获取项目列表
func (client *DBClient) ListProjectByOrgID(orgID uint64) ([]model.Project, error) {
	var projects []model.Project
	if err := client.Where("org_id = ?", orgID).Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

// ListProjectByCluster 根据clusterName 获取项目列表
func (client *DBClient) ListProjectByCluster(clusterName string) ([]model.Project, error) {
	var projects []model.Project
	if err := client.Where("cluster_config LIKE ?", "%"+clusterName+"%").Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

// UpdateProjectQuota 更新项目配额
func (client *DBClient) UpdateProjectQuota(clusterName string, cpuOverSellChangeRatio float64) error {
	return client.Debug().Model(model.Project{}).
		Where("cluster_config LIKE ?", "%"+clusterName+"%").
		Update("cpu_quota", gorm.Expr("cpu_quota * ?", cpuOverSellChangeRatio)).Error
}

type ProjectID struct {
	ProjectID string `json:"project_id"`
}

// GetJoinedProjectNumByUserID get projects by userID and orgID
func (client *DBClient) GetJoinedProjectNumByUserID(userID string, orgID string) (int, []string, error) {
	var total int
	var proIDS []ProjectID
	res := make([]string, 0)
	if err := client.Model(&model.Member{}).
		Where("user_id = ? and org_id = ? and scope_type = \"?\"", userID, orgID, apistructs.ProjectScopeType).
		Select("project_id").Find(&proIDS).Offset(0).Limit(-1).Count(&total).Error; err != nil {
		return total, res, err
	}
	for _, v := range proIDS {
		res = append(res, v.ProjectID)
	}
	return total, res, nil
}

// GetProjectIDListByStates get states by projectID list
func (client *DBClient) GetProjectIDListByStates(req apistructs.IssuePagingRequest, projectIDList []uint64) (int, []model.Project, error) {
	var (
		total int
		res   []model.Project
	)
	sql := client.Table("ps_group_projects").Where("id in (select distinct project_id from dice_issues where deleted = 0 and project_id in (?) and assignee IN (?) and state IN (?) and type IN(?) )", projectIDList, req.Assignees, req.State, req.Type).
		Order("name")
	offset := (req.PageNo - 1) * req.PageSize
	if err := sql.Offset(offset).Limit(req.PageSize).Find(&res).Error; err != nil {
		return total, res, err
	}
	if err := sql.Count(&total).Error; err != nil {
		return total, res, err
	}
	return total, res, nil
}
