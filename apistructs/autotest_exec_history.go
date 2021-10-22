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

type AutoTestExecHistoryDto struct {
	ID            uint64         `json:"id"`
	CreatorID     string         `json:"creatorID"`
	ProjectID     uint64         `json:"projectID"`
	SpaceID       uint64         `json:"spaceID"`
	IterationID   uint64         `json:"iterationID"`
	PlanID        uint64         `json:"planID"`
	SceneID       uint64         `json:"sceneID"`
	SceneSetID    uint64         `json:"sceneSetID"`
	StepID        uint64         `json:"stepID"`
	ParentPID     uint64         `json:"parentPID"`
	Type          StepAPIType    `json:"type"`
	Status        PipelineStatus `json:"status"`
	PipelineYml   string         `json:"pipelineYml"`
	ExecuteApiNum int64          `json:"executeApiNum"`
	SuccessApiNum int64          `json:"successApiNum"`
	PassRate      float64        `json:"passRate"`
	ExecuteRate   float64        `json:"executeRate"`
	TotalApiNum   int64          `json:"totalApiNum"`
	ExecuteTime   time.Time      `json:"executeTime"`
	CostTimeSec   int64          `json:"costTimeSec"`
	OrgID         uint64         `json:"orgID"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}

type ExecHistorySceneAvgCostTime struct {
	SceneID uint64  `json:"sceneID" gorm:"scene_id"`
	Avg     float64 `json:"avg" gorm:"avg"`
}

type ExecHistorySceneStatusCount struct {
	SceneID      uint64  `json:"sceneID" gorm:"scene_id"`
	SuccessCount uint64  `json:"successCount" gorm:"success_count"`
	FailCount    uint64  `json:"failCount" gorm:"fail_count"`
	FailRate     float64 `json:"failRate"`
}

type ExecHistorySceneApiStatusCount struct {
	SceneID      uint64  `json:"sceneID" gorm:"scene_id"`
	SuccessCount uint64  `json:"successCount" gorm:"success_count"`
	FailCount    uint64  `json:"failCount" gorm:"fail_count"`
	PassRate     float64 `json:"passRate"`
}

type ExecHistoryApiAvgCostTime struct {
	StepID uint64  `json:"stepID" gorm:"step_id"`
	Avg    float64 `json:"avg" gorm:"avg"`
}

type ExecHistoryApiStatusCount struct {
	StepID       uint64  `json:"stepID" gorm:"step_id"`
	SuccessCount uint64  `json:"successCount" gorm:"success_count"`
	FailCount    uint64  `json:"failCount" gorm:"fail_count"`
	FailRate     float64 `json:"failRate"`
}
