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

type PipelineStage struct {
	ID         uint64 `json:"id" xorm:"pk autoincr"`
	PipelineID uint64 `json:"pipelineID"`

	Name   string                    `json:"name"`
	Extra  PipelineStageExtra        `json:"extra" xorm:"json"`
	Status apistructs.PipelineStatus `json:"status"`

	CostTimeSec int64     `json:"costTimeSec"`
	TimeBegin   time.Time `json:"timeBegin"`                  // 执行开始时间
	TimeEnd     time.Time `json:"timeEnd"`                    // 执行结束时间
	TimeCreated time.Time `json:"timeCreated" xorm:"created"` // 记录创建时间
	TimeUpdated time.Time `json:"timeUpdated" xorm:"updated"` // 记录更新时间
}

type PipelineStageExtra struct {
	PreStage   *PreStageSimple `json:"preStage,omitempty"`
	StageOrder int             `json:"stageOrder"` // 0,1,2,...
}

type PreStageSimple struct {
	ID     uint64                    `json:"id"`
	Status apistructs.PipelineStatus `json:"preStageStatus,omitempty"`
}

func (ps *PipelineStage) TableName() string {
	return "pipeline_stages"
}

func (ps *PipelineStage) Convert2DTO() *apistructs.PipelineStageDTO {
	if ps == nil {
		return nil
	}
	return &apistructs.PipelineStageDTO{
		ID:          ps.ID,
		PipelineID:  ps.PipelineID,
		Name:        ps.Name,
		Status:      ps.Status,
		CostTimeSec: ps.CostTimeSec,
		TimeBegin:   ps.TimeBegin,
		TimeEnd:     ps.TimeEnd,
		TimeCreated: ps.TimeCreated,
		TimeUpdated: ps.TimeUpdated,
	}
}
