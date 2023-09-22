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

type PersonalPerformanceRequest struct {
	Start       string                  `json:"start"`
	End         string                  `json:"end"`
	OrgID       uint64                  `json:"orgID"`
	UserID      uint64                  `json:"userID"`
	ProjectIDs  []uint64                `json:"projectIDs"`
	Operations  []ReportFilterOperation `json:"operations"`
	LabelQuerys []ReportLabelOperation  `json:"labelQuerys"` // deliberately use labelQuerys instead of labelQueries
}

type PersonalContributorRequest struct {
	Start          string   `json:"start"`
	End            string   `json:"end"`
	OrgID          uint64   `json:"orgID"`
	UserID         uint64   `json:"userID"`
	UserEmail      string   `json:"userEmail"`
	ProjectIDs     []uint64 `json:"projectIDs"`
	GroupByProject bool     `json:"groupByProject"`
}

type FuncPointTrendRequest struct {
	Start          string   `json:"start"`
	End            string   `json:"end"`
	OrgID          uint64   `json:"orgID"`
	UserID         uint64   `json:"userID"`
	ProjectIDs     []uint64 `json:"projectIDs"`
	GroupByProject bool     `json:"groupByProject"`
}
