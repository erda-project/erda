package dbclient

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Monitor struct {
	Id                 int       `gorm:"column:id" db:"id" json:"id" form:"id"`
	MonitorId          string    `gorm:"column:monitor_id" db:"monitor_id" json:"monitor_id" form:"monitor_id"`
	TerminusKey        string    `gorm:"column:terminus_key" db:"terminus_key" json:"terminus_key" form:"terminus_key"`
	TerminusKeyRuntime string    `gorm:"column:terminus_key_runtime" db:"terminus_key_runtime" json:"terminus_key_runtime" form:"terminus_key_runtime"`
	Workspace          string    `gorm:"column:workspace" db:"workspace" json:"workspace" form:"workspace"`
	RuntimeId          string    `gorm:"column:runtime_id" db:"runtime_id" json:"runtime_id" form:"runtime_id"`
	RuntimeName        string    `gorm:"column:runtime_name" db:"runtime_name" json:"runtime_name" form:"runtime_name"`
	ApplicationId      string    `gorm:"column:application_id" db:"application_id" json:"application_id" form:"application_id"`
	ApplicationName    string    `gorm:"column:application_name" db:"application_name" json:"application_name" form:"application_name"`
	ProjectId          string    `gorm:"column:project_id" db:"project_id" json:"project_id" form:"project_id"`
	ProjectName        string    `gorm:"column:project_name" db:"project_name" json:"project_name" form:"project_name"`
	OrgId              string    `gorm:"column:org_id" db:"org_id" json:"org_id" form:"org_id"`
	OrgName            string    `gorm:"column:org_name" db:"org_name" json:"org_name" form:"org_name"`
	ClusterId          string    `gorm:"column:cluster_id" db:"cluster_id" json:"cluster_id" form:"cluster_id"`
	ClusterName        string    `gorm:"column:cluster_name" db:"cluster_name" json:"cluster_name" form:"cluster_name"`
	Config             string    `gorm:"column:config" db:"config" json:"config" form:"config"`
	CallbackUrl        string    `gorm:"column:callback_url" db:"callback_url" json:"callback_url" form:"callback_url"`
	Version            string    `gorm:"column:version" db:"version" json:"version" form:"version"`
	Plan               string    `gorm:"column:plan" db:"plan" json:"plan" form:"plan"`
	IsDelete           int8      `gorm:"column:is_delete" db:"is_delete" json:"is_delete" form:"is_delete"`
	Created            time.Time `gorm:"column:created" db:"created" json:"created" form:"created"`
	Updated            time.Time `gorm:"column:updated" db:"updated" json:"updated" form:"updated"`
}

func (Monitor) TableName() string { return "sp_monitor" }

type MonitorDB struct {
	*gorm.DB
}

func (db *DBClient) GetMonitorByProjectIdAndWorkspace(projectID int64, workspace string) (*Monitor, error) {
	monitor := Monitor{}
	err := db.Where("`project_id` = ?", projectID).Where("`workspace` = ?", workspace).Find(&monitor).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &monitor, nil
}
