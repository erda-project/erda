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
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/internal/tools/orchestrator/spec"
)

// RuntimeVPA define K8s VPA object for runtime's service
type RuntimeVPA struct {
	ID                     string    `json:"id" gorm:"size:36"`
	CreatedAt              time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt              time.Time `json:"updated_at"gorm:"column:updated_at"`
	RuleName               string    `json:"rule_name"`
	RuleNameSpace          string    `json:"rule_namespace" gorm:"column:rule_namespace"`
	OrgID                  uint64    `json:"org_id" gorm:"not null"`
	OrgName                string    `json:"org_name"`
	OrgDisPlayName         string    `json:"org_display_name" gorm:"column:org_display_name"`
	ProjectID              uint64    `json:"project_id" gorm:"not null"`
	ProjectName            string    `json:"project_name"`
	ProjectDisplayName     string    `json:"proj_display_name" gorm:"column:proj_display_name"`
	ApplicationID          uint64    `json:"application_id" gorm:"not null"`
	ApplicationName        string    `json:"application_name"`
	ApplicationDisPlayName string    `json:"app_display_name" gorm:"column:app_display_name"`
	RuntimeID              uint64    `json:"runtime_id" gorm:"not null"`
	RuntimeName            string    `json:"runtime_name"`
	ClusterName            string    `json:"cluster_name"` // target k8s cluster name
	Workspace              string    `json:"workspace" gorm:"column:workspace"`
	UserID                 string    `json:"user_id"`   // user ID
	UserName               string    `json:"user_name"` // user name
	NickName               string    `json:"nick_name"` // user nick name
	ServiceName            string    `json:"service_name"`
	Rules                  string    `json:"rules" gorm:"type:text"`
	IsApplied              string    `json:"is_applied" gorm:"column:applied"` // ‘Y’ means vpa rule have applied，‘N’ means vpa rule have canceled
	SoftDeletedAt          uint64    `json:"soft_deleted_at" gorm:"column:soft_deleted_at"`
}

func (RuntimeVPA) TableName() string {
	return "erda_v2_runtime_vpa_rule"
}

// RuntimeVPAContainerRecommendation define VPA objects for runtime's service
type RuntimeVPAContainerRecommendation struct {
	ID                    string    `json:"id" gorm:"size:36"`
	CreatedAt             time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt             time.Time `json:"updated_at"gorm:"column:updated_at"`
	RuleName              string    `json:"rule_name"`
	RuleID                string    `json:"rule_id"`
	RuleNameSpace         string    `json:"rule_namespace" gorm:"column:rule_namespace"`
	OrgID                 uint64    `json:"org_id" gorm:"not null"`
	OrgName               string    `json:"org_name"`
	ProjectID             uint64    `json:"project_id" gorm:"not null"`
	ProjectName           string    `json:"project_name"`
	ApplicationID         uint64    `json:"application_id" gorm:"not null"`
	ApplicationName       string    `json:"application_name"`
	RuntimeID             uint64    `json:"runtime_id" gorm:"not null"`
	RuntimeName           string    `json:"runtime_name"`
	Workspace             string    `json:"workspace" gorm:"column:workspace"`
	ClusterName           string    `json:"cluster_name"` // target k8s cluster name
	ServiceName           string    `json:"service_name"`
	ContainerName         string    `json:"container_name"`
	LowerCPURequest       float64   `json:"lower_bound_cpu_request" gorm:"lower_cpu_request"`
	LowerMemoryRequest    float64   `json:"lower_bound_memory_request" gorm:"lower_memory_request"`
	UpperCPURequest       float64   `json:"upper_bound_cpu_request" gorm:"upper_cpu_request"`
	UpperMemoryRequest    float64   `json:"upper_bound_memory_request" gorm:"upper_memory_request"`
	TargetCPURequest      float64   `json:"target_cpu_request" gorm:"target_cpu_request"`                  // real cpu value apply to pod
	TargetMemoryRequest   float64   `json:"target_memory_request" gorm:"target_memory_request"`            // real memory value apply to pod
	UncappedCPURequest    float64   `json:"uncapped_target_cpu_request" gorm:"uncapped_cpu_request"`       // no limits target cpu value
	UncappedMemoryRequest float64   `json:"uncapped_target_memory_request" gorm:"uncapped_memory_request"` // no limits target memory value
	SoftDeletedAt         uint64    `json:"soft_deleted_at" gorm:"column:soft_deleted_at"`
}

func (RuntimeVPAContainerRecommendation) TableName() string {
	return "erda_v2_runtime_vpa_recommendation"
}

func (db *DBClient) CreateRuntimeVPA(runtimeVPA *RuntimeVPA) error {
	return db.Save(runtimeVPA).Error
}

func (db *DBClient) UpdateRuntimeVPA(runtimeVPA *RuntimeVPA) error {
	if err := db.Model(&RuntimeVPA{}).Where("id = ?", runtimeVPA.ID).Update(runtimeVPA).Error; err != nil {
		return errors.Wrapf(err, "failed to update runtime vpa rule, id: %v", runtimeVPA.ID)
	}
	return nil
}

// if not found, return (nil, error)
func (db *DBClient) GetRuntimeVPAByServices(id spec.RuntimeUniqueId, services []string) ([]RuntimeVPA, error) {
	var runtimeVPAs []RuntimeVPA
	if len(services) > 0 {
		if err := db.
			Where("application_id = ? AND workspace = ? AND runtime_name = ? AND service_name in (?)", id.ApplicationId, id.Workspace, id.Name, services).
			Find(&runtimeVPAs).Error; err != nil {
			return nil, errors.Wrapf(err, "failed to get runtime vpa rule for runtime %+v for services: %v", id, services)
		}
	} else {
		if err := db.
			Where("application_id = ? AND workspace = ? AND runtime_name = ? ", id.ApplicationId, id.Workspace, id.Name).
			Find(&runtimeVPAs).Error; err != nil {
			return nil, errors.Wrapf(err, "failed to get runtime vpa rule for runtime: %+v", id)
		}
	}
	return runtimeVPAs, nil
}

func (db *DBClient) DeleteRuntimeVPAByRuleId(ruleId string) error {
	if err := db.
		Where("id = ?", ruleId).
		Delete(&RuntimeVPA{}).Error; err != nil {
		return errors.Wrapf(err, "failed to delete runtime vpa rule for rule id: %v", ruleId)
	}
	return nil
}

func (db *DBClient) GetRuntimeVPARuleByRuleId(ruleId string) (*RuntimeVPA, error) {
	var runtimeVPA RuntimeVPA
	if err := db.
		Where("id = ?", ruleId).
		Find(&runtimeVPA).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to get runtime vpa rule for rule id: %v", ruleId)
	}
	return &runtimeVPA, nil
}

func (db *DBClient) GetRuntimeVPARulesByRuntimeId(runtimeId uint64) ([]RuntimeVPA, error) {
	var runtimeVPAs []RuntimeVPA
	if err := db.
		Where("runtime_id = ?", runtimeId).
		Find(&runtimeVPAs).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to get runtime vpa rule for runtime_id: %v", runtimeId)
	}
	return runtimeVPAs, nil
}

// if not found, return (nil, error)
func (db *DBClient) GetRuntimeVPARecommendationsByServices(runtimeId uint64, services []string) ([]RuntimeVPAContainerRecommendation, error) {
	var runtimeVPARecommendations []RuntimeVPAContainerRecommendation
	if len(services) > 0 {
		if err := db.
			Where("runtime_id = ? AND service_name in (?)", runtimeId, services).
			Find(&runtimeVPARecommendations).Error; err != nil {
			return nil, errors.Wrapf(err, "failed to get runtime vpa recommendations for runtimeId %v for services: %v", runtimeId, services)
		}
	} else {
		if err := db.
			Where("runtime_id = ?", runtimeId).
			Find(&runtimeVPARecommendations).Error; err != nil {
			return nil, errors.Wrapf(err, "failed to get runtime vpa recommendations for runtimeId %v", runtimeId)
		}
	}
	return runtimeVPARecommendations, nil
}
