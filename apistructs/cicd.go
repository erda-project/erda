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

// CICDPipelineListRequest /api/cicds 获取 pipeline 列表
type CICDPipelineListRequest struct {
	Branches string `schema:"branches"`
	Sources  string `schema:"sources"`
	YmlNames string `schema:"ymlNames"`
	Statuses string `schema:"statuses"`
	AppID    uint64 `schema:"appID"`
	PageNum  int    `schema:"pageNum"`
	PageSize int    `schema:"pageSize"`
}

// CICDPipelineYmlListRequest /api/cicds/actions/pipelineYmls 获取 pipeline yml列表
type CICDPipelineYmlListRequest struct {
	AppID  int64  `schema:"appID"`
	Branch string `schema:"branch"`
}

// CICDPipelineYmlListResponse
type CICDPipelineYmlListResponse struct {
	Data []string `json:"data"`
}
