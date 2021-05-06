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

// PipelineBase represents `dice_pipeline_reports` table.
type PipelineReport struct {
	ID         uint64 `xorm:"pk autoincr"`
	PipelineID uint64
	Type       apistructs.PipelineReportType
	Meta       apistructs.PipelineReportMeta `xorm:"json"`
	CreatorID  string
	UpdaterID  string
	CreatedAt  time.Time `xorm:"created"`
	UpdatedAt  time.Time `xorm:"updated"`
}

func (*PipelineReport) TableName() string {
	return "dice_pipeline_reports"
}
