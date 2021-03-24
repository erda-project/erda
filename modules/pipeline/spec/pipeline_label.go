package spec

import (
	"time"

	"github.com/erda-project/erda/apistructs"
)

// PipelineLabel 标签
type PipelineLabel struct {
	ID uint64 `json:"id" xorm:"pk autoincr"`

	PipelineSource  apistructs.PipelineSource `json:"pipelineSource"`
	PipelineYmlName string                    `json:"pipelineYmlName"`
	PipelineID      uint64                    `json:"pipelineID"`

	Key   string `json:"key"`
	Value string `json:"value"`

	TimeCreated time.Time `json:"timeCreated" xorm:"created"`
	TimeUpdated time.Time `json:"timeUpdated" xorm:"updated"`
}

func (p PipelineLabel) TableName() string {
	return "pipeline_labels"
}
