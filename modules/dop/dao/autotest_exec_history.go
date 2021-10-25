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

package dao

import (
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

type AutoTestExecHistory struct {
	dbengine.BaseModel

	CreatorID     string
	ProjectID     uint64
	SpaceID       uint64
	IterationID   uint64
	PlanID        uint64
	SceneID       uint64
	SceneSetID    uint64
	StepID        uint64
	ParentPID     uint64
	Type          apistructs.StepAPIType
	Status        apistructs.PipelineStatus
	PipelineYml   string // Used to record the order of scenes,sceneSets and steps
	ExecuteApiNum int64
	SuccessApiNum int64
	PassRate      float64
	ExecuteRate   float64
	TotalApiNum   int64
	ExecuteTime   time.Time
	CostTimeSec   int64
	OrgID         uint64
	TimeBegin     time.Time
	TimeEnd       time.Time
	PipelineID    uint64
}

func (a *AutoTestExecHistory) Convert() apistructs.AutoTestExecHistoryDto {
	return apistructs.AutoTestExecHistoryDto{
		ID:            a.ID,
		CreatorID:     a.CreatorID,
		ProjectID:     a.ProjectID,
		SpaceID:       a.SpaceID,
		IterationID:   a.IterationID,
		PlanID:        a.PlanID,
		SceneID:       a.SceneID,
		SceneSetID:    a.SceneSetID,
		StepID:        a.StepID,
		ParentPID:     a.ParentPID,
		Type:          a.Type,
		Status:        a.Status,
		PipelineYml:   a.PipelineYml,
		ExecuteApiNum: a.ExecuteApiNum,
		SuccessApiNum: a.SuccessApiNum,
		PassRate:      a.PassRate,
		ExecuteRate:   a.ExecuteRate,
		TotalApiNum:   a.TotalApiNum,
		ExecuteTime:   a.ExecuteTime,
		CostTimeSec:   a.CostTimeSec,
		OrgID:         a.OrgID,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
		PipelineID:    a.PipelineID,
	}
}

func (AutoTestExecHistory) TableName() string {
	return "dice_autotest_exec_history"
}

// ListAutoTestExecHistory .
func (client *DBClient) ListAutoTestExecHistory(timeStart, timeEnd string, planIDs ...uint64) ([]AutoTestExecHistory, error) {
	var list []AutoTestExecHistory
	db := client.Debug().Model(&AutoTestExecHistory{}).
		Where("plan_id IN (?)", planIDs)
	if timeStart != "" {
		db = db.Where("execute_time >= ?", timeStart)
	}

	if timeEnd != "" {
		db = db.Where("execute_time <= ?", timeEnd)
	}
	err := db.Find(&list).Order("execute_time ASC").Error
	return list, err
}
