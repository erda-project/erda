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
