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

package bundle

import (
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/modules/pkg/diceworkspace"
	"github.com/erda-project/erda/pkg/http/httputil"
)

// GetProjectBranchRules 查询项目分支规则
func (b *Bundle) GetProjectBranchRules(projectId uint64) ([]*apistructs.BranchRule, error) {
	return b.GetBranchRules(apistructs.ProjectScope, projectId)
}

// GetAppBranchRules 查询应用分支规则
func (b *Bundle) GetAppBranchRules(appId uint64) ([]*apistructs.BranchRule, error) {
	return b.GetBranchRules(apistructs.AppScope, appId)
}

// GetBranchRules 查询分支规则
func (b *Bundle) GetBranchRules(scopeType apistructs.ScopeType, scopeID uint64) ([]*apistructs.BranchRule, error) {
	host, err := b.urls.DOP()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var fetchResp apistructs.QueryBranchRuleResponse
	resp, err := hc.Get(host).Path("/api/branch-rules").
		Param("scopeId", strconv.FormatUint(scopeID, 10)).
		Param("scopeType", string(scopeType)).
		Header(httputil.InternalHeader, "bundle").Do().JSON(&fetchResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !fetchResp.Success {
		return nil, toAPIError(resp.StatusCode(), fetchResp.Error)
	}

	return fetchResp.Data, nil
}

func (b *Bundle) GetAllValidBranchWorkspace(appId uint64) ([]apistructs.ValidBranch, error) {
	host, err := b.urls.DOP()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var fetchResp apistructs.PipelineAppAllValidBranchWorkspaceResponse
	resp, err := hc.Get(host).Path("/api/branch-rules/actions/app-all-valid-branch-workspaces").
		Param("appId", strconv.FormatUint(appId, 10)).
		Header(httputil.InternalHeader, "bundle").Do().JSON(&fetchResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !fetchResp.Success {
		return nil, toAPIError(resp.StatusCode(), fetchResp.Error)
	}

	return fetchResp.Data, nil
}

func (b *Bundle) GetBranchWorkspaceConfig(appId uint64, branch string) (*apistructs.ValidBranch, error) {
	app, err := b.GetApp(appId)
	if err != nil {
		return nil, err
	}
	return b.GetBranchWorkspaceConfigByProject(app.ProjectID, branch)
}

func (b *Bundle) GetPermissionByGitReference(branch *apistructs.ValidBranch) string {
	resource := "normalBranch"
	if branch.IsProtect {
		resource = "protectedBranch"
	}
	return resource
}

func (b *Bundle) GetBranchWorkspaceConfigByProject(projectID uint64, branch string) (*apistructs.ValidBranch, error) {
	rules, err := b.GetProjectBranchRules(projectID)
	if err != nil {
		return nil, err
	}
	validBranch := diceworkspace.GetValidBranchByGitReference(branch, rules)
	return validBranch, nil
}
