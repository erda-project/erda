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
	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	cronpb "github.com/erda-project/erda-proto-go/core/pipeline/cron/pb"
)

type PipelineCronPagingRequest struct {
	AllSources bool
	Sources    []PipelineSource `schema:"source"`  // ?source=cdp-dev&source=cdp-test
	YmlNames   []string         `schema:"ymlName"` // ?ymlName=11&ymlName=22

	PageSize int
	PageNo   int
}

type PipelineCronPagingResponse struct {
	Header
	Data *PipelineCronPagingResponseData `json:"data"`
}

type PipelineCronPagingResponseData struct {
	Total int64                  `json:"total"`
	Data  []*cronpb.Cron `json:"data,omitempty"`
}

//type PipelineCronDTO struct {
//	ID          uint64    `json:"id"`
//	TimeCreated time.Time `json:"timeCreated"` // 记录创建时间
//	TimeUpdated time.Time `json:"timeUpdated"` // 记录更新时间
//
//	ApplicationID   uint64     `json:"applicationID"`
//	Branch          string     `json:"branch"`
//	CronExpr        string     `json:"cronExpr"`
//	CronStartTime   *time.Time `json:"cronStartTime"`
//	PipelineYmlName string     `json:"pipelineYmlName"` // 一个分支下可以有多个 pipeline 文件，每个分支可以有单独的 cron 逻辑
//	BasePipelineID  uint64     `json:"basePipelineID"`  // 用于记录最开始创建出这条 cron 记录的 pipeline id
//	Enable          *bool      `json:"enable"`          // 1 true, 0 false
//}

type PipelineCronCreateRequest struct {
	PipelineCreateRequest *basepb.PipelineCreateRequest `json:"pipelineCreateRequest"`
}

type PipelineCronCreateResponse struct {
	Header
	Data uint64 `json:"data"` // cronID
}

type PipelineCronDeleteResponse struct {
	Header
}
