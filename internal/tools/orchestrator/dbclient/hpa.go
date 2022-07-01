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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	hpatypes "github.com/erda-project/erda/internal/tools/orchestrator/components/horizontalpodscaler/types"
	"github.com/erda-project/erda/internal/tools/orchestrator/spec"
)

// RuntimeHPA define KEDA ScaledObjects for runtime's service
type RuntimeHPA struct {
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
	IsApplied              string    `json:"is_applied" gorm:"column:applied"` // ‘Y’ means hpa rule have applied，‘N’ means hpa rule have canceled
	SoftDeletedAt          uint64    `json:"soft_deleted_at" gorm:"column:soft_deleted_at"`
}

func (RuntimeHPA) TableName() string {
	return "erda_v2_runtime_hpa"
}

type HPAEventInfo struct {
	ID            string    `json:"id" gorm:"size:36"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt     time.Time `json:"updated_at"gorm:"column:updated_at"`
	RuntimeID     uint64    `json:"runtime_id" gorm:"not null"`
	OrgID         uint64    `json:"org_id" gorm:"not null"`
	OrgName       string    `json:"org_name"`
	ServiceName   string    `json:"service_name"`
	Event         string    `json:"event" gorm:"type:text"`
	SoftDeletedAt uint64    `json:"soft_deleted_at" gorm:"column:soft_deleted_at"`
}

type EventDetail struct {
	LastTimestamp metav1.Time `json:"lastTimestamp,omitempty"`
	Type          string      `json:"type,omitempty"`
	Reason        string      `json:"reason,omitempty"`
	Message       string      `json:"message,omitempty"`
}

func (HPAEventInfo) TableName() string {
	return "erda_v2_runtime_hpa_events"
}

func (db *DBClient) CreateRuntimeHPA(runtimeHPA *RuntimeHPA) error {
	return db.Save(runtimeHPA).Error
}

func (db *DBClient) UpdateRuntimeHPA(runtimeHPA *RuntimeHPA) error {
	if err := db.Model(&RuntimeHPA{}).Where("id = ?", runtimeHPA.ID).Update(runtimeHPA).Error; err != nil {
		return errors.Wrapf(err, "failed to update runtime hpa rule, id: %v", runtimeHPA.ID)
	}
	return nil
}

// if not found, return (nil, error)
func (db *DBClient) GetRuntimeHPAByServices(id spec.RuntimeUniqueId, services []string) ([]RuntimeHPA, error) {
	var runtimeHPAs []RuntimeHPA
	if len(services) > 0 {
		if err := db.
			Where("application_id = ? AND workspace = ? AND runtime_name = ? AND service_name in (?)", id.ApplicationId, id.Workspace, id.Name, services).
			Find(&runtimeHPAs).Error; err != nil {
			return nil, errors.Wrapf(err, "failed to get runtime hpa rule for runtime %+v for services: %v", id, services)
		}
	} else {
		if err := db.
			Where("application_id = ? AND workspace = ? AND runtime_name = ? ", id.ApplicationId, id.Workspace, id.Name).
			Find(&runtimeHPAs).Error; err != nil {
			return nil, errors.Wrapf(err, "failed to get runtime hpa rule for runtime: %+v", id)
		}
	}
	return runtimeHPAs, nil
}

func (db *DBClient) DeleteRuntimeHPAByRuleId(ruleId string) error {
	if err := db.
		Where("id = ?", ruleId).
		Delete(&RuntimeHPA{}).Error; err != nil {
		return errors.Wrapf(err, "failed to delete runtime hpa rule for rule id: %v", ruleId)
	}
	return nil
}

func (db *DBClient) GetRuntimeHPARuleByRuleId(ruleId string) (*RuntimeHPA, error) {
	var runtimeHPA RuntimeHPA
	if err := db.
		Where("id = ?", ruleId).
		Find(&runtimeHPA).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to get runtime hpa rule for rule id: %v", ruleId)
	}
	return &runtimeHPA, nil
}

func (db *DBClient) GetRuntimeHPARulesByRuntimeId(runtimeId uint64) ([]RuntimeHPA, error) {
	var runtimeHPAs []RuntimeHPA
	if err := db.
		Where("runtime_id = ?", runtimeId).
		Find(&runtimeHPAs).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to get runtime hpa rule for runtime_id: %v", runtimeId)
	}
	return runtimeHPAs, nil
}

func (db *DBClient) CreateHPAEventInfo(hpaEvent *HPAEventInfo) error {
	return db.Save(hpaEvent).Error
}

// if not found, return (nil, error)
func (db *DBClient) GetRuntimeHPAEventsByServices(runtimeId uint64, services []string) ([]HPAEventInfo, error) {
	var hpaEvents []HPAEventInfo
	if len(services) > 0 {
		//  select * from erda_v2_runtime_hpa_events  where runtime_id = '143' AND service_name in ('go-demo','abc') order by updated_at desc limit 20;
		for _, svc := range services {
			var hpaEventsForService []HPAEventInfo
			if svc != "" {
				if err := db.
					Where("runtime_id = ? AND service_name = ?", runtimeId, svc).Order("updated_at desc").Limit(hpatypes.ErdaHPARecentlyEventsMaxToListForServiceDefault).
					Find(&hpaEventsForService).Error; err != nil {
					return nil, errors.Wrapf(err, "failed to get runtime hpa events for runtimeId %+v for service: %v", runtimeId, svc)
				}
				hpaEvents = append(hpaEvents, hpaEventsForService...)
			}
		}
	} else {
		if err := db.
			Where("runtime_id = ?", runtimeId).Order("updated_at desc").Limit(hpatypes.ErdaHPARecentlyEventsMaxToListForRuntimeDefault).
			Find(&hpaEvents).Error; err != nil {
			return nil, errors.Wrapf(err, "failed to get runtime hpa events for runtimeId: %+v", runtimeId)
		}
	}
	return hpaEvents, nil
}

func (db *DBClient) DeleteRuntimeHPAEventsByRuleId(ruleId string) error {
	if err := db.
		Where("id = ?", ruleId).
		Delete(&HPAEventInfo{}).Error; err != nil {
		return errors.Wrapf(err, "failed to delete runtime hpa events for rule_id: %v", ruleId)
	}
	return nil
}
