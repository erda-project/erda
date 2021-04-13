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

package api

type CreateTagRequest struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Ref     string `json:"ref"`
}

type CreateBranchRequest struct {
	Name string `json:"name"`
	Ref  string `json:"ref"`
}

type CreateRepoRequest struct {
	OrgID       int64             `json:"org_id"`
	ProjectID   int64             `json:"project_id"`
	AppID       int64             `json:"app_id"`
	OrgName     string            `json:"org_name"`
	ProjectName string            `json:"project_name"`
	AppName     string            `json:"app_name"`
	HostMode    string            `json:"host_mode"` //selfhost|external
	Config      map[string]string `json:"config"`
}

type CreateRepoResponseData struct {
	ID int64 `json:"id"`

	// 仓库相对路径
	RepoPath string `json:"repo_path"`
}

type MergeTemplatesResponseData struct {
	Branch string   `json:"branch"`
	Path   string   `json:"path"`
	Names  []string `json:"names"`
}

// 分页查询
type PagingRequest struct {
	// +optional default 1
	PageNo int `json:"PageNo"`
	// +optional default 10
	PageSize int `json:"PageSize"`
}
