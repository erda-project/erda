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

import (
	"encoding/json"
	"time"
)

// Request for API: `GET /api/deployments`
type DeploymentListRequest struct {
	PageInfo

	// 应用实例 ID
	RuntimeID uint64 `query:"runtimeId"`

	// Org ID, 获取 'orgid' 下的所有 runtime 的 deployments
	OrgID uint64 `query:"orgId"`

	// 通过 Status 过滤，不传为默认不过滤
	StatusIn string `query:"statusIn"`
}

// Response for API: `GET /api/deployments`
type DeploymentListResponse struct {
	Header
	UserInfoHeader
	Data *DeploymentListData `json:"data"`
}

type DeploymentListData struct {
	Total int           `json:"total"`
	List  []*Deployment `json:"list"`
}

type Deployment struct {
	ID             uint64           `json:"id"`
	RuntimeID      uint64           `json:"runtimeId"`
	BuildID        uint64           `json:"buildId"`
	ReleaseID      string           `json:"releaseId"`
	ReleaseName    string           `json:"releaseName"`
	Type           string           `json:"type"`
	Status         DeploymentStatus `json:"status"`
	Phase          DeploymentPhase  `json:"phase"`
	Step           DeploymentPhase  `json:"step"` // Deprecated: use phase instead
	FailCause      string           `json:"failCause"`
	Outdated       bool             `json:"outdated"`
	NeedApproval   bool             `json:"needApproval"`
	ApprovedByUser string           `json:"approvedByUser"`
	ApprovedAt     *time.Time       `json:"approvedAt"`
	ApprovalStatus string           `json:"approvalStatus"`
	ApprovalReason string           `json:"approvalReason"`

	Operator       string     `json:"operator"`
	OperatorName   string     `json:"operatorName"`   // Deprecated
	OperatorAvatar string     `json:"operatorAvatar"` // Deprecated
	CreatedAt      time.Time  `json:"createdAt"`
	FinishedAt     *time.Time `json:"finishedAt"`
	RollbackFrom   uint64     `json:"rollbackFrom"`
}

type DeploymentDetailListResponse struct {
	Header
	UserInfoHeader
	Data *DeploymentDetailListData `json:"data"`
}
type DeploymentWithDetail struct {
	Deployment
	RuntimeName     string `json:"runtimeName"`
	ApplicationName string `json:"applicationName"`
	ApplicationID   uint64 `json:"applicationId"`
	ProjectName     string `json:"projectName"`
	ProjectID       uint64 `json:"projectId"`
	BranchName      string `json:"branchName"`
	CommitID        string `json:"commitId"`
	CommitMessage   string `json:"commitMessage"`
}

type DeploymentDetailListData struct {
	Total int                     `json:"total"`
	List  []*DeploymentWithDetail `json:"list"`
}

type DeploymentCancelRequest struct {
	RuntimeID json.Number `json:"runtimeId"`
	Operator  string      `json:"operator"`
}

type DeploymentCancelResponse struct {
	Header
	// no data
}

type DeploymentApproveRequest struct {
	ID     uint64 `json:"id"`
	Reject bool   `json:"reject"`
	Reason string `json:"reason"`
}
type DeploymentApproveResponse struct {
	Header
}

type DeployStagesAddonsRequest struct {
}
type DeployStagesServicesRequest struct {
}
type DeployStagesDomainsRequest struct {
}
