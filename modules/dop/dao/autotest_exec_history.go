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
	ParentID      uint64
	Type          apistructs.StepAPIType
	Status        apistructs.PipelineStatus
	PipelineYml   string // Used to record the order of scenes,sceneSets and steps
	ExecuteApiNum int64
	SuccessApiNum int64
	PassRate      float64
	ExecuteRate   float64
	TotalApiNum   int64
	ExecuteTime   time.Time
}

func (AutoTestExecHistory) TableName() string {
	return "dice_autotest_exec_history"
}
