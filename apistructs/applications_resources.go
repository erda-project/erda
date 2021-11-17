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

type ApplicationsResourcesRequest struct {
	UserID string
	OrgID uint64
	ProjectID uint64
	Query *ApplicationsResourceQuery
}

type ApplicationsResourceQuery struct {
	ApplicationsIDs []uint64
	OwnerIDs        []uint64
	OrderBy         []string
	PageNo          uint64
	PageSize        uint64
}

type ApplicationsResourcesResponse struct {
	Total uint64 `json:"total"`
	List  []ApplicationsResourcesItem
}

type ApplicationsResourcesItem struct {
	Id                int    `json:"id"` // the application primary
	Name              string `json:"name"`
	DisplayName       string `json:"displayName"`
	OwnerUserID       int    `json:"ownerUserID"`
	OwnerUserName     string `json:"ownerUserName"`
	OwnerUserNickname string `json:"ownerUserNickname"`
	PodsCount         int    `json:"runtimesCount"`
	CPURequest        int    `json:"cpuRequest"`
	MemRequest        int    `json:"memRequest"`
	ProdCPURequest    int    `json:"prodCPURequest"`
	ProdMemRequest    int    `json:"prodMemRequest"`
	ProdPodsCount     int    `json:"prodRuntimesCount"`
	StagingCPURequest int    `json:"stagingCPURequest"`
	StagingMemRequest int    `json:"stagingMemRequest"`
	StagingPodsCount  int    `json:"stagingRuntimesCount"`
	TestCPURequest    int    `json:"testCPURequest"`
	TestMemRequest    int    `json:"testMemRequest"`
	TestPodsCount     int    `json:"testRuntimesCount"`
	DevCPURequest     int    `json:"devCPURequest"`
	DevMemRequest     int    `json:"devMemRequest"`
	DevPodsCount      int    `json:"devRuntimesCount"`
}
