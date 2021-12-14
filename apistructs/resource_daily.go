//  Copyright (c) 2021 Terminus, Inc.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package apistructs

import "time"

// ProjectResourceDailyModel is the model cmp_prject_resource_daily
type ProjectResourceDailyModel struct {
	ID        uint64    `json:"id" gorm:"id"`
	CreatedAt time.Time `json:"created_at" gorm:"created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"updated_at"`

	ProjectID          uint64 `json:"project_id" gorm:"project_id"`
	ProjectName        string `json:"project_name" gorm:"project_name"`
	ProjectDisplayName string `json:"project_display_name" gorm:"project_display_name"`

	OwnerUserID       uint64 `json:"owner_user_id" gorm:"owner_user_id"`
	OwnerUserName     string `json:"owner_user_name" gorm:"owner_user_name"`
	OwnerUserNickname string `json:"owner_user_nickname" gorm:"owner_user_nickname"`

	ClusterName string `json:"cluster_name" gorm:"cluster_name"`
	CPUQuota    uint64 `json:"cpu_quota" gorm:"cpu_quota"`
	CPURequest  uint64 `json:"cpu_request" gorm:"cpu_request"`
	MemQuota    uint64 `json:"mem_quota" gorm:"mem_quota"`
	MemRequest  uint64 `json:"mem_request" gorm:"mem_request"`
}

func (m ProjectResourceDailyModel) TableName() string {
	return "cmp_project_resource_daily"
}

func (m ProjectResourceDailyModel) CreatedDay() string {
	return m.CreatedAt.Format("2006-01-02")
}

func (m ProjectResourceDailyModel) UpdatedDay() string {
	return m.UpdatedAt.Format("2006-01-02")
}

//ClusterResourceDailyModel is the model cmp_cluster_resource_daily
type ClusterResourceDailyModel struct {
	ID        uint64    `json:"id" gorm:"id"`
	CreatedAt time.Time `json:"created_at" gorm:"created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"updated_at"`

	ClusterName  string `json:"cluster_name" gorm:"cluster_name"`
	CPUTotal     uint64 `json:"cpu_total" gorm:"cpu_total"`
	CPURequested uint64 `json:"cpu_requested" gorm:"cpu_requested"`
	MemTotal     uint64 `json:"mem_total" gorm:"mem_total"`
	MemRequested uint64 `json:"mem_requested" gorm:"mem_requested"`
}

func (m ClusterResourceDailyModel) TableName() string {
	return "cmp_cluster_resource_daily"
}

func (m ClusterResourceDailyModel) CreatedDay() string {
	return m.CreatedAt.Format("2006-01-02")
}

func (m ClusterResourceDailyModel) UpdatedDay() string {
	return m.UpdatedAt.Format("2006-01-02")
}

// ApplicationResourceDailyModel is the model cmp_application_resource_daily
type ApplicationResourceDailyModel struct {
	ID        uint64    `json:"id" gorm:"id"`
	CreatedAt time.Time `json:"created_at" gorm:"created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"updated_at"`

	ProjectID              uint64 `json:"project_id" gorm:"project_id"`
	ApplicationID          uint64 `json:"application_id" gorm:"application_id"`
	ApplicationName        string `gorm:"application_name"`
	ApplicationDisplayName string `gorm:"application_display_name"`

	ProdCPURequest uint64 `json:"prod_cpu_request" gorm:"prod_cpu_request"`
	ProdMemRequest uint64 `json:"prod_mem_request" gorm:"prod_mem_request"`
	ProdPodsCount  uint64 `json:"prod_pods_count" gorm:"prod_pods_count"`

	StagingCPURequest uint64 `json:"staging_cpu_request" gorm:"staging_cpu_request"`
	StagingMemRequest uint64 `json:"staging_mem_request" gorm:"staging_mem_request"`
	StagingPodsCount  uint64 `json:"staging_pods_count" gorm:"staging_pods_count"`

	TestCPURequest uint64 `json:"test_cpu_request" gorm:"test_cpu_request"`
	TestMemRequest uint64 `json:"test_mem_request" gorm:"test_mem_request"`
	TestPodsCount  uint64 `json:"test_pods_count" gorm:"test_pods_count"`

	DevCPURequest uint64 `json:"dev_cpu_request" gorm:"dev_cpu_request"`
	DevMemRequest uint64 `json:"dev_mem_request" gorm:"dev_mem_request"`
	DevPodsCount  uint64 `json:"dev_pods_count" gorm:"dev_pods_count"`
}

func (ApplicationResourceDailyModel) TableName() string {
	return "cmp_application_resource_daily"
}
