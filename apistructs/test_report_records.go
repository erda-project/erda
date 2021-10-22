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
	"strconv"
	"time"
)

type TestReportRecord struct {
	IdentityInfo

	ID           uint64         `json:"id"`
	ProjectID    uint64         `json:"projectID"`
	IterationID  uint64         `json:"iterationID"`
	CreatorID    string         `json:"creatorID"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	Name         string         `json:"name"`
	QualityScore float64        `json:"qualityScore"`
	ReportData   TestReportData `json:"reportData"`
}

type TestReportData struct {
	IssueDashboard *ComponentProtocol `json:"issue-dashboard,omitempty"`
	TestDashboard  *ComponentProtocol `json:"test-dashboard,omitempty"`
}

type TestReportRecordListRequest struct {
	IdentityInfo

	ID            uint64 `json:"id"`
	Name          string `json:"name"`
	ProjectID     uint64 `json:"projectID"`
	IterationID   uint64 `json:"iterationID"`
	GetReportData bool   `json:"getReportData"`
	OrderBy       string `json:"orderBy"`
	Asc           bool   `json:"asc"`
	PageNo        uint64 `json:"pageNo"`
	PageSize      uint64 `json:"pageSize"`
}

func (req *TestReportRecordListRequest) URLQueryString() map[string][]string {
	query := make(map[string][]string)
	if req.Name != "" {
		query["name"] = []string{req.Name}
	}
	if req.ProjectID != 0 {
		query["projectID"] = []string{strconv.FormatInt(int64(req.ProjectID), 10)}
	}
	if req.IterationID != 0 {
		query["iterationID"] = []string{strconv.FormatInt(int64(req.IterationID), 10)}
	}
	query["getReportData"] = []string{strconv.FormatBool(req.GetReportData)}
	if req.OrderBy != "" {
		query["orderBy"] = []string{req.OrderBy}
	}
	query["asc"] = []string{strconv.FormatBool(req.Asc)}
	query["pageNo"] = []string{strconv.FormatInt(int64(req.PageNo), 10)}
	query["pageSize"] = []string{strconv.FormatInt(int64(req.PageSize), 10)}
	return query
}

type TestReportRecordData struct {
	Total uint64             `json:"total"`
	List  []TestReportRecord `json:"list"`
}

type CreateTestReportRecordResponse struct {
	Header
	Id uint64 `json:"id"`
}

type ListTestReportRecordResponse struct {
	Header
	Data *TestReportRecordData `json:"data"`
}

type GetTestReportRecordResponse struct {
	Header
	Data TestReportRecord `json:"data"`
}
