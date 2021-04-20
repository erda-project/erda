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

// PipelineBase represents `pipeline_bases` table.
type PipelineBase struct {
	ID uint64 `json:"id" xorm:"pk autoincr"`

	PipelineSource  apistructs.PipelineSource `json:"pipelineSource"`
	PipelineYmlName string                    `json:"pipelineYmlName"`

	ClusterName string `json:"clusterName,omitempty"`

	Status apistructs.PipelineStatus `json:"status,omitempty"`

	Type        apistructs.PipelineType        `json:"type,omitempty"`
	TriggerMode apistructs.PipelineTriggerMode `json:"triggerMode,omitempty"`

	// 定时相关信息
	// +optional
	CronID *uint64 `json:"cronID,omitempty"`

	// Snippet
	IsSnippet        bool    `json:"isSnippet"`
	ParentPipelineID *uint64 `json:"parentPipelineID,omitempty"`
	ParentTaskID     *uint64 `json:"parentTaskID,omitempty"`

	// CostTimeSec 总耗时(秒)
	CostTimeSec int64 `json:"costTimeSec,omitempty"` // pipeline 总耗时/秒
	// TimeBegin 执行开始时间
	TimeBegin *time.Time `json:"timeBegin,omitempty"` // 执行开始时间
	// TimeEnd 执行结束时间
	TimeEnd *time.Time `json:"timeEnd,omitempty"` // 执行结束时间

	TimeCreated *time.Time `json:"timeCreated,omitempty" xorm:"created"`
	TimeUpdated *time.Time `json:"timeUpdated,omitempty" xorm:"updated"`
}

func (*PipelineBase) TableName() string {
	return "pipeline_bases"
}
