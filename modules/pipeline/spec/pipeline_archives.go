// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package spec

import (
	"time"

	"github.com/erda-project/erda/apistructs"
)

// PipelineArchive pipeline 归档表
type PipelineArchive struct {
	ID          uint64    `json:"id" xorm:"pk autoincr"`
	TimeCreated time.Time `json:"timeCreated" xorm:"created"`
	TimeUpdated time.Time `json:"timeUpdated" xorm:"updated"`

	PipelineID      uint64                    `json:"pipelineID"`
	PipelineSource  apistructs.PipelineSource `json:"pipelineSource"`
	PipelineYmlName string                    `json:"pipelineYmlName"`
	Status          apistructs.PipelineStatus `json:"status"`

	// DiceVersion record the dice version when archived,
	// it will impact `content` field unmarshal method
	DiceVersion string                 `json:"diceVersion"`
	Content     PipelineArchiveContent `json:"content" xorm:"json"`
}

// PipelineArchiveContent contains:
// - pipelines
// - pipeline_labels
// - pipeline_stages
// - pipeline_tasks
type PipelineArchiveContent struct {
	Pipeline        Pipeline         `json:"pipeline"`
	PipelineLabels  []PipelineLabel  `json:"pipelineLabels"`
	PipelineStages  []PipelineStage  `json:"pipelineStages"`
	PipelineTasks   []PipelineTask   `json:"pipelineTasks"`
	PipelineReports []PipelineReport `json:"pipelineReports"`
}

func (*PipelineArchive) TableName() string {
	return "pipeline_archives"
}
