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
