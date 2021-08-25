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

type WorkbenchRequest struct {
	OrgID     uint64 `json:"orgID"`
	PageNo    int    `json:"pageNo"`
	PageSize  int    `json:"pageSize"`
	IssueSize int    `json:"issueSize"`
}

type WorkbenchResponse struct {
	Header
	Data WorkbenchResponseData `json:"data"`
}

type WorkbenchResponseData struct {
	TotalProject int                     `json:"totalProject"`
	TotalIssue   int                     `json:"totalIssue"`
	List         []*WorkbenchProjectItem `json:"list"`
}

type WorkbenchProjectItem struct {
	ProjectDTO          ProjectDTO `json:"projectDTO"`
	TotalIssueNum       int        `json:"totalIssueNum"`
	UnSpecialIssueNum   int        `json:"unSpecialIssueNum"`
	ExpiredIssueNum     int        `json:"expiredIssueNum"`
	ExpiredOneDayNum    int        `json:"expiredOneDayNum"`
	ExpiredTomorrowNum  int        `json:"expiredTomorrowNum"`
	ExpiredSevenDayNum  int        `json:"expiredSevenDayNum"`
	ExpiredThirtyDayNum int        `json:"expiredThirtyDayNum"`
	FeatureDayNum       int        `json:"featureDayNum"`
	IssueList           []Issue    `json:"issueList"`
}
