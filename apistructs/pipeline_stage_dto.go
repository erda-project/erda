package apistructs

import (
	"time"
)

type PipelineStageDTO struct {
	ID         uint64 `json:"id"`
	PipelineID uint64 `json:"pipelineID"`

	Name   string         `json:"name"`
	Status PipelineStatus `json:"status"`

	CostTimeSec int64     `json:"costTimeSec"`
	TimeBegin   time.Time `json:"timeBegin"`
	TimeEnd     time.Time `json:"timeEnd"`
	TimeCreated time.Time `json:"timeCreated"`
	TimeUpdated time.Time `json:"timeUpdated"`
}
