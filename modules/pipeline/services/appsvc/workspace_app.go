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

package appsvc

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pkg/diceworkspace"
	"github.com/erda-project/erda/pkg/strutil"
)

// WorkspaceApp 包括所有 pipeline 创建需要的应用信息
type WorkspaceApp struct {
	ID            uint64                   `json:"ID"`
	Name          string                   `json:"name"`
	OrgID         uint64                   `json:"orgId"`
	OrgName       string                   `json:"orgName"`
	ProjectID     uint64                   `json:"projectID"`
	ProjectName   string                   `json:"projectName"`
	GitRepo       string                   `json:"gitRepo"`
	GitRepoAbbrev string                   `json:"gitRepoAbbrev"`
	Branch        string                   `json:"branch"`
	Workspace     apistructs.DiceWorkspace `json:"workspace"`
	ClusterName   string                   `json:"clusterName"`
}

func (s *AppSvc) GetWorkspaceApp(appID uint64, branch string) (*WorkspaceApp, error) {
	var result = WorkspaceApp{ID: appID}

	// get app
	app, err := s.bdl.GetApp(appID)
	if err != nil {
		return nil, err
	}
	result.Name = app.Name
	result.OrgID = app.OrgID
	result.ProjectID = app.ProjectID
	result.GitRepo = app.GitRepo
	result.GitRepoAbbrev = app.GitRepoAbbrev
	result.Branch = branch

	// get org
	org, err := s.bdl.GetOrg(app.OrgID)
	if err != nil {
		return nil, err
	}
	result.OrgName = org.Name

	// get project
	pj, err := s.bdl.GetProject(app.ProjectID)
	if err != nil {
		return nil, err
	}
	result.ProjectName = pj.Name

	rules, err := s.bdl.GetProjectBranchRules(app.ProjectID)
	if err != nil {
		return nil, err
	}
	// get workspace by branch
	wsByBranch, err := diceworkspace.GetByGitReference(branch, rules)
	if err != nil {
		return nil, err
	}
	var foundWorkspace bool
	for ws, clusterName := range pj.ClusterConfig {
		if strutil.Equal(ws, string(wsByBranch), true) {
			result.Workspace = wsByBranch
			result.ClusterName = clusterName
			foundWorkspace = true
			break
		}
	}
	if !foundWorkspace {
		return nil, errors.Errorf("failed to found corresponding workspace in application info, "+
			"branch: %s, workspace: %s", branch, wsByBranch)
	}

	return &result, nil
}

func (app *WorkspaceApp) GenerateLabels() map[string]string {
	labels := make(map[string]string)

	// org
	labels[apistructs.LabelOrgID] = fmt.Sprintf("%d", app.OrgID)
	labels[apistructs.LabelOrgName] = app.OrgName

	// project
	labels[apistructs.LabelProjectID] = fmt.Sprintf("%d", app.ProjectID)
	labels[apistructs.LabelProjectName] = app.ProjectName

	// app
	labels[apistructs.LabelAppID] = fmt.Sprintf("%d", app.ID)
	labels[apistructs.LabelAppName] = app.Name

	// workspace
	labels[apistructs.LabelDiceWorkspace] = app.Workspace.String()

	// branch
	labels[apistructs.LabelBranch] = app.Branch

	for k, v := range labels {
		if v == "" {
			delete(labels, k)
		}
	}
	return labels
}

func (app *WorkspaceApp) GenerateV1UniquePipelineYmlName(originPipelineYmlPath string) string {
	// 若 originPipelineYmlPath 已经符合生成规则，则直接返回
	ss := strutil.Split(originPipelineYmlPath, "/", true)
	if len(ss) > 2 {
		appID, _ := strconv.ParseUint(ss[0], 10, 64)
		workspace := ss[1]
		if appID == app.ID && workspace == app.Workspace.String() {
			return originPipelineYmlPath
		}
	}
	return fmt.Sprintf("%d/%s/%s/%s", app.ID, app.Workspace.String(), app.Branch, originPipelineYmlPath)
}
