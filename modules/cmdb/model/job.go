package model

import "time"

// Jobs 运行的job信息
type Jobs struct {
	ID              int64     `json:"id" gorm:"primary_key"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	OrgID           uint64    `gorm:"index:org_id"`
	ProjectID       uint64
	ApplicationID   uint64
	PipelineID      uint64
	TaskID          uint64
	QueueTimeSec    int64 // 排队耗时
	CostTimeSec     int64 // 任务耗时
	ProjectName     string
	ApplicationName string
	TaskName        string
	Status          string
	Env             string
	ClusterName     string
	TaskType        string
	UserID          string
	Extra           ExtraJob `json:"extra"`
}

// 额外字段预留
type ExtraJob struct{}

// TableName 设置模型对应数据库表名称
func (Jobs) TableName() string {
	return "cm_jobs"
}
