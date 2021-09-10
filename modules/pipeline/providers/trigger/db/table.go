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

import (
	"time"

	"github.com/erda-project/erda/apistructs"
)

// PipelineBase represents `pipeline_triggers` table.
type PipelineTrigger struct {
	ID                   uint64                    `json:"id" xorm:"pk autoincr"`
	Event                string                    `json:"event" xorm:"event"`
	PipelineSource       apistructs.PipelineSource `json:"pipelineSource" xorm:"pipeline_source"`
	PipelineYmlName      string                    `json:"pipelineYmlName" xorm:"pipeline_yml_name"`
	PipelineDefinitionID uint64                    `json:"pipelineDefinitionID" xorm:"pipeline_definition_id"`
	Filter               map[string]string         `json:"filter" xorm:"filter"` // TODO change to query once in the database
	CreatedAt            *time.Time                `json:"createdAt,omitempty" xorm:"created_at created"`
	UpdatedAt            *time.Time                `json:"updatedAt,omitempty" xorm:"updated_at updated"`
}

func (*PipelineTrigger) TableName() string {
	return "pipeline_triggers"
}
