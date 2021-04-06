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
	"fmt"
	"time"

	"github.com/erda-project/erda/pkg/strutil"
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
	PipelineReportTypeBasic   PipelineReportType = "basic"
	PipelineReportTypeAPITest PipelineReportType = "api-test"
)

// PipelineReportMeta 流水线报告元数据，前端根据该数据拼装报告详情界面
type PipelineReportMeta map[string]interface{}

// PipelineReportCreateRequest 报告创建请求
type PipelineReportCreateRequest struct {
	PipelineID uint64             `json:"pipelineID"`
	Type       PipelineReportType `json:"type"`
	Meta       PipelineReportMeta `json:"meta"`

	IdentityInfo
}

func (req PipelineReportCreateRequest) BasicValidate() error {
	if req.PipelineID == 0 {
		return fmt.Errorf("missing pipelineID")
	}
	if err := strutil.Validate(string(req.Type), strutil.MinLenValidator(1), strutil.MaxLenValidator(32)); err != nil {
		return fmt.Errorf("invalid type: %v", err)
	}
	return nil
}

type PipelineReportCreateResponse struct {
	Header
	Data *PipelineReport `json:"data"`
}

type PipelineReportSetGetResponse struct {
	Header
	Data *PipelineReportSet `json:"data"`
}

type PipelineReportSetPagingRequest struct {
	PipelineIDs []uint64             `schema:"-"`
	Sources     []PipelineSource     `schema:"source"`
	Types       []PipelineReportType `schema:"type"`

	/////////////////////////
	// pipeline 分页查询参数 //
	/////////////////////////
	// labels
	// &mustMatchLabel=key2=value3
	MustMatchLabelsQueryParams []string `schema:"mustMatchLabel"`

	// times
	// 开始执行时间 左闭区间
	StartTimeBeginTimestamp int64 `schema:"startTimeBeginTimestamp"`
	// 开始执行时间 右闭区间
	EndTimeBeginTimestamp int64 `schema:"endTimeBeginTimestamp"`
	// 创建时间 左闭区间
	StartTimeCreatedTimestamp int64 `schema:"startTimeCreatedTimestamp"`
	// 创建时间 右闭区间
	EndTimeCreatedTimestamp int64 `schema:"endTimeCreatedTimestamp"`

	PageNum  int `schema:"pageNum"`
	PageSize int `schema:"pageSize"`
}

type PipelineReportSetPagingResponse struct {
	Header
	Data *PipelineReportSetPagingResponseData `json:"data"`
}

type PipelineReportSetPagingResponseData struct {
	Total     int                 `json:"total"`
	Pipelines []PipelineReportSet `json:"reportSets"`
}
