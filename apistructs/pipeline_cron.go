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

type PipelineCronPagingRequest struct {
	AllSources bool             `schema:"allSources"`
	Sources    []PipelineSource `schema:"source"`  // ?source=cdp-dev&source=cdp-test
	YmlNames   []string         `schema:"ymlName"` // ?ymlName=11&ymlName=22
	Enable     *bool            `schema:"enable"`

	PageSize int `schema:"pageSize"`
	PageNo   int `schema:"pageNo"`
}

type PipelineCronPagingResponse struct {
	Header
	Data *PipelineCronPagingResponseData `json:"data"`
}

type PipelineCronPagingResponseData struct {
	Total int64              `json:"total"`
	Data  []*PipelineCronDTO `json:"data,omitempty"`
}

type PipelineCronDTO struct {
	ID          uint64    `json:"id"`
	TimeCreated time.Time `json:"timeCreated"` // 记录创建时间
	TimeUpdated time.Time `json:"timeUpdated"` // 记录更新时间

	ApplicationID          uint64     `json:"applicationID"`
	Branch                 string     `json:"branch"`
	CronExpr               string     `json:"cronExpr"`
	CronStartTime          *time.Time `json:"cronStartTime"`
	PipelineYmlName        string     `json:"pipelineYmlName"` // 一个分支下可以有多个 pipeline 文件，每个分支可以有单独的 cron 逻辑
	BasePipelineID         uint64     `json:"basePipelineID"`  // 用于记录最开始创建出这条 cron 记录的 pipeline id
	Enable                 *bool      `json:"enable"`          // 1 true, 0 false
	PipelineYml            string     `json:"pipelineYml"`
	ConfigManageNamespaces []string   `json:"configManageNamespaces"`
	UserID                 string     `json:"userID"`
	OrgID                  uint64     `json:"orgID"`
}

type PipelineCronCreateRequest struct {
	PipelineCreateRequest PipelineCreateRequestV2 `json:"pipelineCreateRequest"`
}

type PipelineCronCreateResponse struct {
	Header
	Data uint64 `json:"data"` // cronID
}

type PipelineCronDeleteResponse struct {
	Header
}

type PipelineCronUpdateRequest struct {
	ID                     uint64   `json:"id"`
	PipelineYml            string   `json:"pipelineYml"`
	CronExpr               string   `json:"cronExpr"`
	ConfigManageNamespaces []string `json:"configManageNamespaces"`
}

type PipelineCronUpdateResponse struct {
	Header
}
