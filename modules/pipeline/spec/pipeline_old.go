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

// PipelineOld
type PipelineOld struct {
	ID uint64 `json:"id,omitempty" xorm:"pk autoincr"`

	// 通过 source + pipelineYmlName 唯一定位
	Source apistructs.PipelineSource `json:"source,omitempty"`
	// 通过 v1 创建的 pipeline，自动生成唯一的 pipelineYmlName
	// 通过 v2 创建的 pipeline，由调用方保证
	PipelineYmlName string `json:"pipelineYmlName,omitempty"`
	PipelineYml     string `json:"pipelineYml,omitempty"`

	// 调度集群
	// +required
	ClusterName string `json:"clusterName,omitempty"`

	// 运行时相关信息
	Type        apistructs.PipelineType        `json:"type,omitempty"`
	TriggerMode apistructs.PipelineTriggerMode `json:"triggerMode,omitempty"`
	Snapshot    Snapshot                       `json:"snapshot,omitempty" xorm:"json"` // 快照
	Progress    float64                        `json:"progress,omitempty" xorm:"-"`    // pipeline 执行进度, eg: 0.8 表示 80%
	Status      apistructs.PipelineStatus      `json:"status,omitempty"`
	Extra       PipelineExtraInfo              `json:"extra,omitempty" xorm:"json"`

	// 时间
	CostTimeSec int64      `json:"costTimeSec,omitempty"`                // pipeline 总耗时/秒
	TimeBegin   *time.Time `json:"timeBegin,omitempty"`                  // 执行开始时间
	TimeEnd     *time.Time `json:"timeEnd,omitempty"`                    // 执行结束时间
	TimeCreated *time.Time `json:"timeCreated,omitempty" xorm:"created"` // 记录创建时间
	TimeUpdated *time.Time `json:"timeUpdated,omitempty" xorm:"updated"` // 记录更新时间

	// 定时相关信息
	// +optional
	CronID *uint64 `json:"cronID,omitempty"`

	// deprecated
	BasePipelineID uint64 `json:"basePipelineID,omitempty"` // 该字段用来分页展示时 group 分组，相同 BasePipelineID 的数据会被折叠成一条，通过执行记录来跳转

	// 应用相关信息
	// +optional
	OrgID           uint64 `json:"orgID,omitempty"`
	OrgName         string `json:"orgName,omitempty"` // tag schedule
	ProjectID       uint64 `json:"projectID,omitempty"`
	ProjectName     string `json:"projectName,omitempty"` // tag schedule
	ApplicationID   uint64 `json:"applicationID,omitempty"`
	ApplicationName string `json:"applicationName,omitempty"`

	// 分支相关信息
	// +optional
	PipelineYmlSource apistructs.PipelineYmlSource `json:"pipelineYmlSource,omitempty"` // yml 文件来源
	Branch            string                       `json:"branch,omitempty"`
	Commit            string                       `json:"commit,omitempty"`
	CommitDetail      apistructs.CommitDetail      `json:"commitDetail,omitempty" xorm:"json"`
}

func (*PipelineOld) TableName() string {
	return "pipelines"
}
