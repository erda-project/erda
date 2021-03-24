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
