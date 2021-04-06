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

type BranchRule struct {
	ID        int64     `json:"id"`
	ScopeType ScopeType `json:"scopeType"`
	ScopeID   int64     `json:"scope_id"`
	Desc      string    `json:"desc"`
	// 分支规则 eg:master,feature/*
	Rule              string `json:"rule"`
	IsProtect         bool   `json:"isProtect"`
	IsTriggerPipeline bool   `json:"isTriggerPipeline"`
	// project级别
	NeedApproval bool `json:"needApproval"`
	// 通过分支创建的流水线环境
	Workspace string `json:"workspace"`
	// 制品可部署的环境
	ArtifactWorkspace string `json:"artifactWorkspace"`
}
type QueryBranchRuleRequest struct {
	ProjectID int64 `query:"projectId"`
	AppID     int64 `query:"appId"`
}

type QueryBranchRuleResponse struct {
	Header
	Data []*BranchRule `json:"data"`
}

type CreateBranchRuleRequest struct {
	ScopeType         ScopeType `json:"scopeType"`
	ScopeID           int64     `json:"scopeId"`
	Rule              string    `json:"rule"`
	IsProtect         bool      `json:"isProtect"`
	NeedApproval      bool      `json:"needApproval"`
	IsTriggerPipeline bool      `json:"isTriggerPipeline"`
	Workspace         string    `json:"workspace"`
	ArtifactWorkspace string    `json:"artifactWorkspace"`
	Desc              string    `json:"desc"`
}

type CreateBranchRuleResponse struct {
	Header
	Data *BranchRule `json:"data"`
}

type UpdateBranchRuleRequest struct {
	ID                int64  `json:"-"`
	Rule              string `json:"rule"`
	IsProtect         bool   `json:"isProtect"`
	NeedApproval      bool   `json:"needApproval"`
	IsTriggerPipeline bool   `json:"isTriggerPipeline"`
	Desc              string `json:"desc"`
	Workspace         string `json:"workspace"`
	ArtifactWorkspace string `json:"artifactWorkspace"`
}

type UpdateBranchRuleResponse struct {
	Header
	Data *BranchRule `json:"data"`
}

type DeleteBranchRuleResponse struct {
	Header
	Data *BranchRule `json:"data"`
}
