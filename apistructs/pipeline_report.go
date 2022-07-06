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

package apistructs

import (
	"time"
)

const (
	PipelineReportEventMetaKey = "event"
	PipelineReportLoopMetaKey  = "task-loop"
)

// PipelineReportSet 流水线报告集，一条流水线可能会有多个报告，称为报告集
type PipelineReportSet struct {
	PipelineID uint64           `json:"pipelineID"`
	Reports    []PipelineReport `json:"reports"`
}

// PipelineReport 流水线报告
type PipelineReport struct {
	ID         uint64             `json:"id"`
	PipelineID uint64             `json:"pipelineID"`
	Type       PipelineReportType `json:"type"`
	Meta       PipelineReportMeta `json:"meta"`
	CreatorID  string             `json:"creatorID"`
	UpdaterID  string             `json:"updaterID"`
	CreatedAt  time.Time          `json:"createdAt"`
	UpdatedAt  time.Time          `json:"updatedAt"`
}

// PipelineReportType 流水线报告类型
type PipelineReportType string

var (
	PipelineReportTypeBasic        PipelineReportType = "basic"
	PipelineReportTypeAPITest      PipelineReportType = "api-test"
	PipelineReportTypeEvent        PipelineReportType = "event"
	PipelineReportTypeInspect      PipelineReportType = "task-inspect"
	PipelineReportTypeAutotestPlan PipelineReportType = "auto-test-execute-config"
)

// PipelineReportMeta 流水线报告元数据，前端根据该数据拼装报告详情界面
type PipelineReportMeta map[string]interface{}

type PipelineReportSetGetResponse struct {
	Header
	Data *PipelineReportSet `json:"data"`
}
