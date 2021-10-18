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

type GetWorkspaceQuotaRequest struct {
	ProjectID string `json:"projectID"`
	Workspace string `json:"workspace"`
}

type GetWorkspaceQuotaResponse struct {
	Header
	Data WorkspaceQuotaData `json:"data"`
}

type WorkspaceQuotaData struct {
	CPU    int64 `json:"cpu"`
	Memory int64 `json:"memory"`
}
