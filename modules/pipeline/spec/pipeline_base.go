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
