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

package dbclient

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/strutil"
)

type Release struct {
	// ReleaseID 唯一标识release, 创建时由服务端生成
	ReleaseID string `json:"releaseId" gorm:"type:varchar(64);primary_key"`
	// ReleaseName 任意字符串，便于用户识别，最大长度255，必填
	ReleaseName string `json:"releaseName" gorm:"index:idx_release_name;not null"`
	// Desc 详细描述此release功能, 选填
	Desc string `json:"desc" gorm:"type:text"`
	// Dice 资源类型为diceyml时, 存储dice.yml内容, 选填
	Dice string `json:"dice" gorm:"type:text"` // dice.yml
	// Addon 资源类型为addonyml时，存储addon.yml内容，选填
	Addon string `json:"addon" gorm:"type:text"`
	// Labels 用于release分类，描述release类别，map类型, 最大长度1000, 选填
	Labels string `json:"labels" gorm:"type:varchar(1000)"`
	// Version 存储release版本信息, 同一企业同一项目同一应用下唯一，最大长度100，选填
	Version string `json:"version" gorm:"type:varchar(100)"` // 版本，打标签的Release不可删除
	// OrgID 企业标识符，描述release所属企业，选填
	OrgID int64 `json:"orgId" gorm:"index:idx_org_id"` // 所属企业
	// ProjectID 项目标识符，描述release所属项目，选填
	ProjectID int64 `json:"projectId"`
	// ApplicationID 应用标识符，描述release所属应用，选填
	ApplicationID int64 `json:"applicationId"`
	// ProjectName 项目名称，描述release所属项目，选填
	ProjectName string `json:"projectName" gorm:"type:varchar(80)"`
	// ApplicationName 应用名称，描述release所属应用，选填
	ApplicationName string `json:"applicationName" gorm:"type:varchar(80)"`
	// UserID 用户标识符, 描述release所属用户，最大长度50，选填
	UserID string `json:"userId" gorm:"type:varchar(50)"`
	// ClusterName 集群名称，描述release所属集群，最大长度80，选填
	ClusterName string `json:"clusterName" gorm:"type:varchar(80)"` // 所属集群
	// Resources 指定release资源类型及资源存储路径, 为兼容现有diceyml，先选填
	Resources string `json:"resources,omitempty" gorm:"type:text"`
	// Reference release被部署次数，当为0时，可清除
	Reference int64 `json:"reference"` // 被部署次数，当为0时，表示可清除
	// CrossCluster 表示当前 release 是否可以跨集群，无集群限制
	CrossCluster bool `json:"crossCluster"`
	// CreatedAt release创建时间，创建时由服务端生成
	CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt release更新时间, 更新时由服务端更新
	UpdatedAt time.Time `json:"updatedAt"`
}

// Set table name
func (Release) TableName() string {
	return "dice_release"
}

// CreateRelease 创建Release
func (client *DBClient) CreateRelease(release *Release) error {
	return client.Create(release).Error
}

// UpdateRelease 更新Release
func (client *DBClient) UpdateRelease(release *Release) error {
	return client.Save(release).Error
}

// DeleteRelease 删除Release
func (client *DBClient) DeleteRelease(releaseID string) error {
	return client.Where("release_id = ?", releaseID).Delete(&Release{}).Error
}

// GetRelease 获取Release
func (client *DBClient) GetRelease(releaseID string) (*Release, error) {
	var release Release
	if err := client.Where("release_id = ?", releaseID).Find(&release).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, dbengine.ErrNotFound
		}
		return nil, err
	}
	return &release, nil
}

// GetReleasesByParams 根据参数过滤Release
func (client *DBClient) GetReleasesByParams(
	orgID, projectID, applicationID int64,
	keyword, releaseName, branch,
	cluster string, crossCluster *bool, isVersion bool, crossClusterOrSpecifyCluster *string,
	startTime, endTime time.Time, pageNum, pageSize int64) (int64, []Release, error) {

	var releases []Release
	db := client.DB.Debug()
	if orgID > 0 {
		db = db.Where("org_id = ?", orgID)
	}
	if applicationID > 0 {
		db = db.Where("application_id = ?", applicationID)
	}

	if projectID > 0 {
		db = db.Where("project_id = ?", projectID)
	}
	if keyword != "" {
		db = db.Where("release_id LIKE ? or release_name LIKE ? or version LIKE ?", "%"+keyword+"%",
			"%"+keyword+"%", "%"+keyword+"%")
	} else if releaseName != "" {
		db = db.Where("release_name = ?", releaseName)
	}
	if isVersion {
		db = db.Not("version", "")
	}

	if cluster != "" {
		db = db.Where("cluster_name = ?", cluster)
	}
	if crossCluster != nil {
		db = db.Where("cross_cluster = ?", *crossCluster)
	}
	if crossClusterOrSpecifyCluster != nil {
		db = db.Where("(cluster_name = ? AND cross_cluster = 0) OR cross_cluster = 1", *crossClusterOrSpecifyCluster)
	}
	if branch != "" {
		db = db.Where("labels LIKE ?", "%"+fmt.Sprintf("\"gitBranch\":\"%s\"", branch)+"%")
	}

	if !startTime.IsZero() {
		db = db.Where("created_at > ?", startTime)
	}

	if err := db.Where("created_at <= ?", endTime).Order("created_at DESC").Offset((pageNum - 1) * pageSize).
		Limit(pageSize).Find(&releases).Error; err != nil {
		return 0, nil, err
	}

	// 获取匹配搜索结果总量
	var total int64
	if err := db.Where("created_at <= ?", endTime).Model(&Release{}).Count(&total).Error; err != nil {
		return 0, nil, err
	}

	return total, releases, nil
}

// GetReleasesByAppAndVersion 根据 appID & version获取 Release列表
func (client *DBClient) GetReleasesByAppAndVersion(orgID, projectID, appID int64, version string) ([]Release, error) {
	var releases []Release
	if err := client.Where("org_id = ?", orgID).
		Where("project_id = ?", projectID).
		Where("application_id = ?", appID).
		Where("version = ?", version).
		Find(&releases).Error; err != nil {
		return nil, err
	}
	return releases, nil
}

// GetReleaseNamesByApp 根据 appID 获取 releaseName 列表
func (client *DBClient) GetReleaseNamesByApp(orgID, appID int64) ([]string, error) {
	var releaseNames []string
	if orgID == 0 {
		if err := client.Select("release_name").
			Where("application_id = ?", appID).
			Group("release_name").Find(&releaseNames).Error; err != nil {
			return nil, err
		}
	} else {
		if err := client.Select("release_name").
			Where("org_id = ?", orgID).
			Where("application_id = ?", appID).
			Group("release_name").Find(&releaseNames).Error; err != nil {
			return nil, err
		}
	}

	return releaseNames, nil
}

// GetAppIDsByProjectAndVersion 根据projectID & version 获取 appID 列表
func (client *DBClient) GetAppIDsByProjectAndVersion(projectID int64, version string) ([]int64, error) {
	var appIDs []int64
	if err := client.Select([]string{"application_id"}).
		Where("project_id = ?", projectID).
		Where("version LIKE ?", strutil.Concat(version, "%")).
		Group("application_id").Find(&appIDs).Error; err != nil {
		return nil, err
	}
	return appIDs, nil
}

// GetLatestReleaseByAppAndVersion 获取应用下最新release
func (client *DBClient) GetLatestReleaseByAppAndVersion(appID int64, version string) (*Release, error) {
	var release Release
	if err := client.Where("application_id = ?", appID).
		Where("version LIKE ?", strutil.Concat(version, "%")).
		Order("created_at DESC").
		Limit(1).Find(&release).Error; err != nil {
		return nil, err
	}
	return &release, nil
}

// GetUnReferedReleasesBefore 获取给定时间点前未引用的 Release
func (client *DBClient) GetUnReferedReleasesBefore(before time.Time) ([]Release, error) {
	var releases []Release
	if err := client.Where("reference <= ?", 0).Where("updated_at < ?", before).
		Order("updated_at").Find(&releases).Error; err != nil {
		return nil, err
	}
	return releases, nil
}
