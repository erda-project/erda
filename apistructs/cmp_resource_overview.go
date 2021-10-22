//  Copyright (c) 2021 Terminus, Inc.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package apistructs

type ResourceOverviewReportData struct {
	Total int                               `json:"total"`
	List  []*ResourceOverviewReportDataItem `json:"list"`
}

type ResourceOverviewReportDataItem struct {
	ProjectID          int64   `json:"projectID"`
	ProjectName        string  `json:"projectName"`
	ProjectDisplayName string  `json:"projectDisplayName"`
	OwnerUserID        int64   `json:"ownerUserID"`
	OwnerUserName      string  `json:"ownerUserName"`
	OwnerUserNickName  string  `json:"ownerUserNickName"`
	CPUQuota           float64 `json:"cpuQuota"`
	CPUWaterLevel      float64 `json:"cpuWaterLevel"`
	MemQuota           float64 `json:"memQuota"`
	MemWaterLevel      float64 `json:"memWaterLevel"`
	Nodes              float64 `json:"nodes"`
}
