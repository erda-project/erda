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

package apistructs

import "time"

// PagePipeline 用于 流水线分页查询结果，相比 PipelineDTO 删除了许多大字段
type PagePipeline struct {
	ID           uint64            `json:"id"`
	CronID       *uint64           `json:"cronID,omitempty"`
	Commit       string            `json:"commit,omitempty"`
	Source       PipelineSource    `json:"source,omitempty"`
	YmlName      string            `json:"ymlName,omitempty"`
	Extra        PipelineExtra     `json:"extra,omitempty"`
	FilterLabels map[string]string `json:"filterLabels"`
	NormalLabels map[string]string `json:"normalLabels"`

	// 运行时相关信息
	Type        string         `json:"type,omitempty"`
	TriggerMode string         `json:"triggerMode,omitempty"`
	ClusterName string         `json:"clusterName,omitempty"`
	Status      PipelineStatus `json:"status,omitempty"`
	Progress    float64        `json:"progress"` // pipeline 执行进度, eg: 0.8 即 80%

	// 嵌套流水线相关信息
	IsSnippet        bool    `json:"isSnippet"`
	ParentPipelineID *uint64 `json:"parentPipelineID,omitempty"`
	ParentTaskID     *uint64 `json:"parentTaskID,omitempty"`

	// 时间
	CostTimeSec int64      `json:"costTimeSec,omitempty"` // pipeline 总耗时/秒
	TimeBegin   *time.Time `json:"timeBegin,omitempty"`   // 执行开始时间
	TimeEnd     *time.Time `json:"timeEnd,omitempty"`     // 执行结束时间
	TimeCreated *time.Time `json:"timeCreated,omitempty"` // 记录创建时间
	TimeUpdated *time.Time `json:"timeUpdated,omitempty"` // 记录更新时间
}
