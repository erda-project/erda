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
	"mime/multipart"
	"time"

	"github.com/pkg/errors"
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

type CodeCoverageStartRequest struct {
	IdentityInfo

	ProjectID uint64 `json:"projectID"`
}

func (req *CodeCoverageStartRequest) Validate() error {
	if req.ProjectID == 0 {
		return errors.New("the projectID is 0")
	}
	return nil
}

type CodeCoverageUpdateRequest struct {
	IdentityInfo

	ID        uint64                `query:"id" schema:"id"`
	Status    string                `query:"status" schema:"status"`
	Msg       string                `query:"msg" schema:"msg"`
	ReportXml *multipart.FileHeader `query:"reportXml" schema:"reportXml"`
	ReportTar *multipart.FileHeader `query:"reportTar" schema:"reportTar"`
}

func (req *CodeCoverageUpdateRequest) Validate() error {
	if req.ID == 0 {
		return errors.New("the ID is 0")
	}
	return nil
}

type CodeCoverageListRequest struct {
	IdentityInfo

	ProjectID uint64                   `json:"projectID"`
	PageNo    uint64                   `json:"pageNo"`
	PageSize  uint64                   `json:"pageSize"`
	TimeBegin string                   `json:"timeBegin"`
	TimeEnd   string                   `json:"timeEnd"`
	Statuses  []CodeCoverageExecStatus `json:"statuses"`
}

func (req *CodeCoverageListRequest) Validate() error {
	if req.ProjectID == 0 {
		return errors.New("the projectID is 0")
	}
	if req.PageNo == 0 {
		req.PageNo = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	return nil
}

type CodeCoverageExecRecordResponse struct {
	Header
	UserInfoHeader
	Data *CodeCoverageExecRecordData `json:"data"`
}

type CodeCoverageExecRecordData struct {
	Total uint64                      `json:"total"`
	List  []CodeCoverageExecRecordDto `json:"list"`
}

type CodeCoverageExecRecordDto struct {
	ID            uint64              `json:"id"`
	ProjectID     uint64              `json:"projectID"`
	Status        string              `json:"status"`
	Msg           string              `json:"msg"`
	Coverage      float64             `json:"coverage"`
	ReportUrl     string              `json:"reportUrl"`
	ReportContent []*CodeCoverageNode `json:"reportContent"`
	StartExecutor string              `json:"startExecutor"`
	EndExecutor   string              `json:"endExecutor"`
	TimeBegin     *time.Time          `json:"timeBegin"`
	TimeEnd       *time.Time          `json:"timeEnd"`
	TimeCreated   time.Time           `json:"timeCreated"`
	TimeUpdated   time.Time           `json:"timeUpdated"`
}

type CodeCoverageNode struct {
	Value []float64           `json:"value"`
	Name  string              `json:"name"`
	Path  string              `json:"path"`
	Nodes []*CodeCoverageNode `json:"children"`
}
