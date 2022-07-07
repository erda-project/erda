// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spec

import (
	"time"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/core/pipeline/report/pb"
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

func (p *PipelineReport) ConvertToPB() (*pb.PipelineReport, error) {
	meta, err := structpb.NewStruct(p.Meta)
	if err != nil {
		return nil, err
	}
	return &pb.PipelineReport{
		ID:         p.ID,
		PipelineID: p.PipelineID,
		Type:       string(p.Type),
		Meta:       meta,
		CreatorID:  p.CreatorID,
		UpdaterID:  p.UpdaterID,
		CreatedAt:  timestamppb.New(p.CreatedAt),
		UpdatedAt:  timestamppb.New(p.UpdatedAt),
	}, nil
}
