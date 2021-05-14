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

// PipelineLabel 标签
type PipelineLabel struct {
	ID uint64 `json:"id" xorm:"pk autoincr"`

	Type     apistructs.PipelineLabelType `json:"type,omitempty"`
	TargetID uint64                       `json:"targetID"`

	PipelineSource  apistructs.PipelineSource `json:"pipelineSource"`
	PipelineYmlName string                    `json:"pipelineYmlName"`

	Key   string `json:"key"`
	Value string `json:"value"`

	TimeCreated time.Time `json:"timeCreated" xorm:"created"`
	TimeUpdated time.Time `json:"timeUpdated" xorm:"updated"`
}

func (p PipelineLabel) TableName() string {
	return "pipeline_labels"
}
