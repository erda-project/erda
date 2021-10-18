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

package apistructs

import "time"

type ProjectQuota struct {
	ID                 uint64    `json:"id" gorm:"id"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"updated_at"`
	CreatedAt          time.Time `json:"created_at" gorm:"created_at"`
	ProjectID          uint64    `json:"project_id" gorm:"project_id"`
	ProjectName        string    `json:"project_name" gorm:"project_name"`
	ProdClusterName    string    `json:"prod_cluster_name" gorm:"prod_cluster_name"`
	StagingClusterName string    `json:"staging_cluster_name" gorm:"staging_cluster_name"`
	TestClusterName    string    `json:"test_cluster_name" gorm:"test_cluster_name"`
	DevClusterName     string    `json:"dev_cluster_name" gorm:"dev_cluster_name"`
	ProdCPUQuota       int64     `json:"prod_cpu_quota" gorm:"prod_cpu_quota"`
	ProdMemQuota       int64     `json:"prod_mem_quota" gorm:"prod_mem_quota"`
	StagingCPUQuota    int64     `json:"staging_cpu_quota" gorm:"staging_cpu_quota"`
	StagingMemQuota    int64     `json:"staging_mem_quota" gorm:"staging_mem_quota"`
	TestCPUQuota       int64     `json:"test_cpu_quota" gorm:"test_cpu_quota"`
	TestMemQuota       int64     `json:"test_mem_quota" gorm:"test_mem_quota"`
	DevCPUQuota        int64     `json:"dev_cpu_quota" gorm:"dev_cpu_quota"`
	DevMemQuota        int64     `json:"dev_mem_quota" gorm:"dev_mem_quota"`
	CreatorID          uint64    `json:"creator_id" gorm:"creator_id"`
	UpdaterID          uint64    `json:"updater_id" gorm:"updater_id"`
}

func (ProjectQuota) TableName() string {
	return "ps_group_projects_quota"
}
