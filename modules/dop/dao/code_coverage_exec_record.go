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

package dao

import (
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

type CodeCoverageExecStatus string

const (
	RunningStatus CodeCoverageExecStatus = "running"
	ReadyStatus   CodeCoverageExecStatus = "ready"
	EndingStatus  CodeCoverageExecStatus = "ending"
	SuccessStatus CodeCoverageExecStatus = "success"
	FailStatus    CodeCoverageExecStatus = "fail"
)

var WorkingStatus = []CodeCoverageExecStatus{RunningStatus, ReadyStatus, EndingStatus}

func (c CodeCoverageExecStatus) String() string {
	return string(c)
}

type CodeCoverageExecRecord struct {
	dbengine.BaseModel

	ProjectID     uint64                         `json:"project_id"`
	Status        CodeCoverageExecStatus         `json:"status"`
	Msg           string                         `json:"msg"`
	Coverage      float64                        `json:"coverage"`
	ReportUrl     string                         `json:"report_url"`
	ReportContent []*apistructs.CodeCoverageNode `json:"report_content"`
	StartExecutor string                         `json:"start_executor"`
	EndExecutor   string                         `json:"end_executor"`
	TimeBegin     *time.Time                     `json:"time_begin"`
	TimeEnd       *time.Time                     `json:"time_end"`
}

func (CodeCoverageExecRecord) TableName() string {
	return "dice_code_coverage_exec_record"
}

func (c *CodeCoverageExecRecord) Covert() *apistructs.CodeCoverageExecRecordDto {
	return &apistructs.CodeCoverageExecRecordDto{
		ID:            c.ID,
		ProjectID:     c.ProjectID,
		Status:        c.Status.String(),
		Msg:           c.Msg,
		Coverage:      c.Coverage,
		ReportUrl:     c.ReportUrl,
		ReportContent: c.ReportContent,
		StartExecutor: c.StartExecutor,
		EndExecutor:   c.EndExecutor,
		TimeBegin:     c.TimeBegin,
		TimeEnd:       c.TimeEnd,
		TimeCreated:   c.CreatedAt,
		TimeUpdated:   c.UpdatedAt,
	}
}

type CodeCoverageExecRecordShort struct {
	dbengine.BaseModel

	ProjectID     uint64                 `json:"project_id"`
	Status        CodeCoverageExecStatus `json:"status"`
	Msg           string                 `json:"msg"`
	Coverage      float64                `json:"coverage"`
	ReportUrl     string                 `json:"report_url"`
	StartExecutor string                 `json:"start_executor"`
	EndExecutor   string                 `json:"end_executor"`
	TimeBegin     *time.Time             `json:"time_begin"`
	TimeEnd       *time.Time             `json:"time_end"`
}

func (CodeCoverageExecRecordShort) TableName() string {
	return "dice_code_coverage_exec_record"
}

func (c *CodeCoverageExecRecordShort) Covert() apistructs.CodeCoverageExecRecordDto {
	return apistructs.CodeCoverageExecRecordDto{
		ID:            c.ID,
		ProjectID:     c.ProjectID,
		Status:        c.Status.String(),
		Msg:           c.Msg,
		Coverage:      c.Coverage,
		ReportUrl:     c.ReportUrl,
		StartExecutor: c.StartExecutor,
		EndExecutor:   c.EndExecutor,
		TimeBegin:     c.TimeBegin,
		TimeEnd:       c.TimeEnd,
		TimeCreated:   c.CreatedAt,
		TimeUpdated:   c.UpdatedAt,
	}
}
