package dao

import (
	"time"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/model"
)

// ListDeploymentsByOrgID 获取指定企业的部署列表
func (client *DBClient) ListDeploymentsByOrgID(param *apistructs.OrgRunningTasksListRequest,
	orgID uint64) (int64, *[]model.Deployments, error) {
	var (
		total       int64
		deployments []model.Deployments
		endedAt     time.Time
		err         error
	)

	db := client.DB.Where("org_id = ?", orgID)
	if param.Cluster != "" {
		db = db.Where("cluster_name = ?", param.Cluster)
	}
	if param.ProjectName != "" {
		db = db.Where("project_name = ?", param.ProjectName)
	}
	if param.AppName != "" {
		db = db.Where("application_name = ?", param.AppName)
	}
	if param.PipelineID != 0 {
		db = db.Where("pipeline_id = ?", param.PipelineID)
	}
	if param.Status != "" {
		db = db.Where("status = ?", param.Status)
	}
	if param.UserID != "" {
		db = db.Where("user_id = ?", param.UserID)
	}
	if param.Env != "" {
		db = db.Where("env = ?", param.Env)
	}

	if param.EndTime == 0 {
		endedAt = time.Now()
	} else {
		endedAt = time.Unix(0, param.EndTime*1000000)
	}
	db = db.Where("created_at < ?", endedAt)
	if param.StartTime != 0 {
		startedAt := time.Unix(0, param.StartTime*1000000)
		db = db.Where("created_at > ?", startedAt)
	}

	if err = db.Order("updated_at DESC").
		Offset((param.PageNo - 1) * param.PageSize).
		Limit(param.PageSize).Find(&deployments).Error; err != nil {
		return 0, nil, err
	}
	// 符合条件的 deployment 总量
	if err = db.Model(&model.Deployments{}).Count(&total).Error; err != nil {
		return 0, nil, err
	}

	return total, &deployments, nil
}

// DeleteDeployment 删除 deployment 信息
func (client *DBClient) DeleteDeployment(orgID string, taskID uint64) error {
	return client.Where("org_id = ?", orgID).Where("task_id = ?", taskID).
		Delete(&model.Deployments{}).Error
}

// GetDeployment 获取部署信息
func (client *DBClient) GetDeployment(orgID string, taskID uint64) []model.Deployments {
	var deployments []model.Deployments
	if err := client.Where("org_id = ?", orgID).
		Where("task_id = ?", taskID).
		Find(&deployments).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil
		}
		return nil
	}
	return deployments
}

// UpdateDeploymentStatus 更新deployment状态
func (client *DBClient) UpdateDeploymentStatus(deployment *model.Deployments) error {
	return client.Save(deployment).Error
}

// CreateDeployment 创建正在运行的deployment
func (client *DBClient) CreateDeployment(deployment *model.Deployments) error {
	return client.Create(deployment).Error
}

// ListExpiredDeployments 列出过期的deployment
func (client *DBClient) ListExpiredDeployments(startTime string) []model.Deployments {
	var deployments []model.Deployments
	if err := client.Where("created_at < ?", startTime).
		Find(&deployments).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil
		}
		return nil
	}
	return deployments
}

// ListRunningDeployments 列出正在运行的deployment
func (client *DBClient) ListRunningDeployments() []model.Deployments {
	var deployments []model.Deployments
	if err := client.Where("status not in (?)",
		[]string{"Success", "AnalyzeFailed", "Timeout", "StopByUser", "Failed", "NoNeedBySystem"}).
		Find(&deployments).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil
		}
		return nil
	}
	return deployments
}
