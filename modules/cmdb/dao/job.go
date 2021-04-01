package dao

import (
	"time"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/model"
)

// ListJobsByOrgID 获取指定企业的job列表
func (client *DBClient) ListJobsByOrgID(param *apistructs.OrgRunningTasksListRequest,
	orgID uint64) (int64, *[]model.Jobs, error) {
	var (
		total   int64
		jobs    []model.Jobs
		endedAt time.Time
		err     error
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
		Limit(param.PageSize).Find(&jobs).Error; err != nil {
		return 0, nil, err
	}

	// 符合条件的 job 总量
	if err = db.Model(&model.Jobs{}).Count(&total).Error; err != nil {
		return 0, nil, err
	}

	return total, &jobs, nil
}

// DeleteJob 删除 Job 信息
func (client *DBClient) DeleteJob(orgID string, taskID uint64) error {
	return client.Where("org_id = ?", orgID).Where("task_id = ?", taskID).
		Delete(&model.Jobs{}).Error
}

// GetJob 获取job信息
func (client *DBClient) GetJob(orgID string, taskID uint64) []model.Jobs {
	var Jobs []model.Jobs
	if err := client.Where("org_id = ?", orgID).
		Where("task_id = ?", taskID).
		Find(&Jobs).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil
		}
		return nil
	}
	return Jobs
}

// UpdateJobStatus 更新Job状态
func (client *DBClient) UpdateJobStatus(Job *model.Jobs) error {
	return client.Save(Job).Error
}

// CreateJob 创建正在运行的Job
func (client *DBClient) CreateJob(Job *model.Jobs) error {
	return client.Create(Job).Error
}

// ListExpiredJobs 列出过期的job
func (client *DBClient) ListExpiredJobs(startTime string) []model.Jobs {
	var jobs []model.Jobs
	if err := client.Where("created_at < ?", startTime).
		Find(&jobs).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil
		}
		return nil
	}
	return jobs
}

// ListRunningJobs 列出正在运行的job
func (client *DBClient) ListRunningJobs() []model.Jobs {
	var jobs []model.Jobs
	if err := client.Where("status not in (?)",
		[]string{"Success", "AnalyzeFailed", "Timeout", "StopByUser", "Failed", "NoNeedBySystem"}).
		Find(&jobs).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil
		}
		return nil
	}
	return jobs
}
