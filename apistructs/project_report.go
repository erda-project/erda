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

type ReportFilterOperation struct {
	Key       string  `json:"key"`
	Val       float64 `json:"val"`
	Operation string  `json:"operation"`
}

type ReportLabelOperation struct {
	Key       string `json:"key"`
	Val       string `json:"val"`
	Operation string `json:"operation"`
}

type ProjectReportRequest struct {
	IsAdmin      bool                    `json:"isAdmin"`
	Start        string                  `json:"start"`
	End          string                  `json:"end"`
	OrgID        uint64                  `json:"orgID"`
	ProjectIDs   []uint64                `json:"projectIDs"`
	IterationIDs []uint64                `json:"iterationIDs"`
	Operations   []ReportFilterOperation `json:"operations"`
	LabelQuerys  []ReportLabelOperation  `json:"labelQuerys"` // deliberately use labelQuerys instead of labelQueries
}
