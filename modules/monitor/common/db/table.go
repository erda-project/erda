// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package db

import "time"

// table name
const (
	InstanceTenantTable = "tb_tmc_instance_tenant"
	MonitorTable        = "sp_monitor"
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
		Id          string `json:"id"`
		MonitorId   string `json:"monitor_id"`
		TerminusKey string `json:"terminus_key"`
		Workspace   string `json:"workspace"`
		ProjectId   string `json:"project_id"`
		ProjectName string `json:"project_name"`
		OrgId       string `json:"org_id"`
		OrgName     string `json:"org_name"`
		ClusterName string `json:"cluster_name"`
		Created     string `json:"created"`
		Updated     string `json:"updated"`
	}
)

func (InstanceTenant) TableName() string {
	return InstanceTenantTable
}

func (Monitor) TableName() string {
	return MonitorTable
}
