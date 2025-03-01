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

package db

import "time"

// table name
const (
	InstanceTenantTable          = "tb_tmc_instance_tenant"
	MonitorTable                 = "sp_monitor"
	SpMonitorConfigRegisterTable = `sp_monitor_config_register`
)

type (
	InstanceTenant struct {
		Id          string    `json:"id"`
		InstanceId  string    `json:"instance_id"`
		Config      string    `json:"config"`
		Options     string    `json:"options"`
		CreateTime  time.Time `json:"create_time"`
		UpdateTime  time.Time `json:"update_time"`
		IsDeleted   string    `json:"is_deleted"`
		TenantGroup string    `json:"tenant_group"`
		Engine      string    `json:"engine"`
		Az          string    `json:"az"`
	}

	Monitor struct {
		Id          string    `gorm:"column:id;primary_key"`
		MonitorId   string    `gorm:"column:monitor_id"`
		TerminusKey string    `gorm:"column:terminus_key"`
		Workspace   string    `gorm:"column:workspace"`
		ProjectId   string    `gorm:"column:project_id"`
		ProjectName string    `gorm:"column:project_name"`
		OrgId       string    `gorm:"column:org_id"`
		OrgName     string    `gorm:"column:org_name"`
		ClusterName string    `gorm:"column:cluster_name"`
		Created     time.Time `gorm:"column:created;default:CURRENT_TIMESTAMP"`
		Updated     time.Time `gorm:"column:updated;default:CURRENT_TIMESTAMP"`
	}

	SpMonitorConfigRegister struct {
		ID         int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
		Scope      string    `gorm:"column:scope;not null" json:"scope"`                  // org
		ScopeID    string    `gorm:"column:scope_id;not null;default:''" json:"scope_id"` // org_id
		Namespace  string    `gorm:"column:namespace;not null" json:"namespace"`          // dev,test,staging,prod,other
		Type       string    `gorm:"column:type;not null" json:"type"`                    // metric、log
		Names      string    `gorm:"column:names;not null" json:"names"`
		Filters    string    `gorm:"column:filters;not null;default:''" json:"filters"`
		Enable     bool      `gorm:"column:enable;not null;default:1" json:"enable"`
		UpdateTime time.Time `gorm:"column:update_time;not null" json:"update_time"`
		Desc       string    `gorm:"column:desc;not null;default:''" json:"desc"`
		Hash       string    `gorm:"column:hash;not null;unique" json:"hash"`
	}
)

func (InstanceTenant) TableName() string {
	return InstanceTenantTable
}

func (Monitor) TableName() string {
	return MonitorTable
}

func (*SpMonitorConfigRegister) TableName() string {
	return SpMonitorConfigRegisterTable
}
