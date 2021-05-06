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

package dbclient

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
)

// AddonInstance addon 实例
type AddonInstance struct {
	ID                  string `gorm:"type:varchar(64);primary_key"` // 主键
	Name                string `gorm:"type:varchar(64)"`             // 用户 dice.yml 指定
	AddonID             string `gorm:"type:varchar(64)"`             // addonID // TODO deprecated
	AddonName           string `gorm:"type:varchar(64)"`             // 应用市场 addon 名称
	Category            string
	Namespace           string
	ScheduleName        string
	Plan                string
	Version             string
	Options             string `gorm:"type:text"`
	Config              string `gorm:"type:text"`
	Label               string
	Status              string
	ShareScope          string // 共享级别 企业/项目/集群/平台
	OrgID               string
	Cluster             string `gorm:"column:az"` // 集群名称
	ProjectID           string
	ApplicationID       string
	AppID               string    `gorm:"column:app_id"`
	Workspace           string    `gorm:"column:env;type:varchar(20)"` // DEV/TEST/STAGING/PROD
	Deleted             string    `gorm:"column:is_deleted"`           // Y: 已删除 N: 未删除
	PlatformServiceType int       `gorm:"type:int(1)"`                 // 服务类型，0：基础addon,1:微服务,2:通用能力
	KmsKey              string    `gorm:"column:kms_key"`
	CreatedAt           time.Time `gorm:"column:create_time"`
	UpdatedAt           time.Time `gorm:"column:update_time"`

	CpuRequest float64
	CpuLimit   float64
	MemRequest int
	MemLimit   int
}

type RemoveAddonID struct {
	ID string `gorm:"column:id"`
}

// TableName 数据库表名
func (AddonInstance) TableName() string {
	return "tb_addon_instance"
}

// UpdateAddonInstanceStatus 根据Id更新信息
func (db *DBClient) UpdateAddonInstanceStatus(ID, status string) error {
	if err := db.Model(&AddonInstance{}).
		Where("id = ?", ID).
		Updates(map[string]interface{}{"status": status}).Error; err != nil {
		return errors.Wrapf(err, "failed to update routing status, instanceID: %v", ID)
	}
	return nil
}

// CreateAddonInstance 创建 addon instance
func (db *DBClient) CreateAddonInstance(instance *AddonInstance) error {
	return db.Create(instance).Error
}

// UpdateAddonInstance 更新 addon instance
func (db *DBClient) UpdateAddonInstance(instance *AddonInstance) error {
	return db.Save(instance).Error
}

// DeleteAddonInstance 删除 addon instance
func (db *DBClient) DeleteAddonInstance(instanceID string) error {
	if err := db.Where("id = ?", instanceID).
		Updates(map[string]interface{}{
			"is_deleted": apistructs.AddonDeleted,
			"status":     apistructs.AddonDetached}).Error; err != nil {
		return err
	}
	return nil
}

// GetAddonInstance 获取 addon instance
func (db *DBClient) GetAddonInstance(id string) (*AddonInstance, error) {
	var instance AddonInstance
	if err := db.Where("id = ?", id).
		Where("is_deleted = ?", apistructs.AddonNotDeleted).
		Find(&instance).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	return &instance, nil
}

// GetConfigCenterAddonInstance 获取配置中心的instance
func (db *DBClient) GetConfigCenterAddonInstance() (*AddonInstance, error) {
	var instance AddonInstance
	if err := db.Where("addon_name = ?", apistructs.AddonConfigCenter).
		Where("is_deleted = ?", apistructs.AddonNotDeleted).
		Find(&instance).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	return &instance, nil
}

// GetInstancesByIDs 根据 ID 查询实例
func (db *DBClient) GetInstancesByIDs(ids []string) (*[]AddonInstance, error) {
	var instances []AddonInstance
	if err := db.Where("id in (?)", ids).
		Where("is_deleted = ?", apistructs.AddonNotDeleted).
		Find(&instances).Error; err != nil {
		return nil, err
	}
	return &instances, nil
}

// ListAddonInstancesByParams 根据参数获取 addon 列表
func (db *DBClient) ListAddonInstancesByParams(orgID uint64, params *apistructs.MiddlewareListRequest) (int, *[]AddonInstance, error) {
	var (
		total     int
		instances []AddonInstance
	)
	dbClient := db.Where("org_id = ?", orgID).Where("platform_service_type = ?", 0)
	if params.ProjectID != 0 {
		dbClient = dbClient.Where("project_id = ?", params.ProjectID)
	}
	if params.AddonName != "" {
		dbClient = dbClient.Where("addon_name = ?", params.AddonName)
	}
	if params.Workspace != "" {
		dbClient = dbClient.Where("env = ?", params.Workspace)
	}
	if params.InstanceID != "" {
		dbClient = dbClient.Where("id = ?", params.InstanceID)
	}
	dbClient = dbClient.Where("is_deleted = ?", apistructs.AddonNotDeleted).
		Where("category not in (?)", []string{apistructs.AddonCustomCategory, apistructs.AddonDiscovery}).
		Where("status in (?)", []apistructs.AddonStatus{apistructs.AddonAttached})

	if err := dbClient.Model(&AddonInstance{}).Count(&total).Error; err != nil {
		return 0, nil, err
	}

	if err := dbClient.Offset((params.PageNo - 1) * params.PageSize).Limit(params.PageSize).Find(&instances).Error; err != nil {
		return 0, nil, err
	}

	return total, &instances, nil
}

// ListAddonInstancesByParamsWithoutPage 根据参数获取 addon 列表
func (db *DBClient) ListAddonInstancesByParamsWithoutPage(orgID uint64, params *apistructs.MiddlewareListRequest) (*[]AddonInstance, error) {
	var instances []AddonInstance

	dbClient := db.Where("org_id = ?", orgID).Where("platform_service_type = ?", 0).Where("category != ?", "discovery")
	if params.ProjectID != 0 {
		dbClient = dbClient.Where("project_id = ?", params.ProjectID)
	}
	if params.AddonName != "" {
		dbClient = dbClient.Where("addon_name = ?", params.AddonName)
	}
	if params.Workspace != "" {
		dbClient = dbClient.Where("env = ?", params.Workspace)
	}
	if params.InstanceID != "" {
		dbClient = dbClient.Where("id = ?", params.InstanceID)
	}
	if params.EndTime != nil {
		dbClient = dbClient.Where("create_time < ?", params.EndTime)
	}
	if err := dbClient.Where("is_deleted = ?", apistructs.AddonNotDeleted).
		Where("status in (?)", []apistructs.AddonStatus{apistructs.AddonAttached, apistructs.AddonAttaching}).
		Find(&instances).Error; err != nil {
		return nil, err
	}

	return &instances, nil
}

//ListAddonInstanceByOrg 根据 orgID 获取实例列表
func (db *DBClient) ListAddonInstanceByOrg(orgID uint64) (*[]AddonInstance, error) {
	var instances []AddonInstance
	if err := db.Where("org_id = ?", orgID).
		Where("is_deleted = ?", apistructs.AddonNotDeleted).
		Where("category != ?", "discovery").
		Where("status in (?)", []apistructs.AddonStatus{apistructs.AddonAttached, apistructs.AddonAttaching}).
		Find(&instances).Error; err != nil {
		return nil, err
	}
	return &instances, nil
}

//ListAddonInstanceByAddonName 根据 addonName 获取实例列表
func (db *DBClient) ListAddonInstanceByAddonName(projectID, workspace, addonName string) (*[]AddonInstance, error) {
	var instances []AddonInstance
	if err := db.Where("project_id = ?", projectID).
		Where("is_deleted = ?", apistructs.AddonNotDeleted).
		Where("env = ?", workspace).
		Where("addon_name = ?", addonName).
		Where("status in (?)", []apistructs.AddonStatus{apistructs.AddonAttached, apistructs.AddonAttaching}).
		Find(&instances).Error; err != nil {
		return nil, err
	}
	return &instances, nil
}

//ListAttachingAddonInstance 查询出所有attaching的addon信息
func (db *DBClient) ListAttachingAddonInstance() (*[]AddonInstance, error) {
	var instances []AddonInstance
	if err := db.Where("is_deleted = ?", apistructs.AddonNotDeleted).
		Where("platform_service_type = ?", 0).
		Where("status in (?)", []apistructs.AddonStatus{apistructs.AddonAttaching}).
		Find(&instances).Error; err != nil {
		return nil, err
	}
	return &instances, nil
}

func (db *DBClient) ListAttachedAddonInstance() ([]AddonInstance, error) {
	var instances []AddonInstance
	if err := db.Where("is_deleted = ?", apistructs.AddonNotDeleted).
		Where("platform_service_type = ?", 0).
		Where("status in (?)", []apistructs.AddonStatus{apistructs.AddonAttached}).
		Find(&instances).Error; err != nil {
		return nil, err
	}
	return instances, nil
}

// ListAddonInstancesByProjectIDs 根据projectIDS列表来返回对应数据
func (db *DBClient) ListAddonInstancesByProjectIDs(projectIDs []uint64, exclude ...string) (*[]AddonInstance, error) {
	var instances []AddonInstance
	exclude_ := []string{"discovery"}
	if len(exclude) != 0 {
		exclude_ = exclude
	}
	if err := db.Where("project_id in (?)", projectIDs).
		Where("is_deleted = ?", apistructs.AddonNotDeleted).
		Where("platform_service_type = ?", 0).
		Where("category not in (?)", exclude_).
		Where("is_deleted = ?", apistructs.AddonNotDeleted).
		Where("status in (?)", []apistructs.AddonStatus{apistructs.AddonAttached}).
		Find(&instances).Error; err != nil {
		return nil, err
	}

	return &instances, nil
}

// CountAddonReferenceByClusterAndOrg 统计集群中addon数量
func (db *DBClient) CountAddonReferenceByClusterAndOrg(clusterName, orgID string) (int, error) {
	var total int

	dbClient := db.Where("org_id = ?", orgID).Where("az = ?", clusterName)
	if err := dbClient.Model(&AddonInstance{}).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// ListNoAttachAddon 查询出所有没有引用关系的addon
func (db *DBClient) ListNoAttachAddon() (*[]RemoveAddonID, error) {
	var result []RemoveAddonID
	if err := db.Table("ps_v2_pre_builds").Raw("select id from (select routing.id,count(att.id) as countNum from tb_addon_instance routing " +
		"left join tb_addon_attachment att on routing.id = att.`instance_id` and att.`is_deleted`=\"N\" where  routing.`is_deleted`=\"N\" " +
		"and routing.`status` in (\"ATTACHED\",\"ATTACHING\") and routing.category != \"custom\" group by routing.id) aa where aa.countNum = 0").Scan(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}
