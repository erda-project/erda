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

package endpoints

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/gittar-adaptor/service/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

func (e *Endpoints) querySnippetYml(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetSnippetYaml.NotLogin().ToResp(), nil
	}
	if !identityInfo.IsInternalClient() {
		return apierrors.ErrGetSnippetYaml.AccessDenied().ToResp(), nil
	}

	var req apistructs.SnippetConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrGetSnippetYaml.InvalidParameter(err).ToResp(), nil
	}

	ymlPath := req.Labels[apistructs.LabelGittarYmlPath]
	if ymlPath == "" {
		return apierrors.ErrGetSnippetYaml.InvalidParameter(fmt.Errorf("labels key %v value is empty", apistructs.LabelGittarYmlPath)).ToResp(), nil
	}

	// get apps by project label
	projectIDString := req.Labels[apistructs.LabelProjectID]
	projectID, err := strconv.Atoi(projectIDString)
	if err != nil {
		return apierrors.ErrGetSnippetYaml.InvalidParameter(err).ToResp(), nil
	}
	project, err := e.bdl.GetProject(uint64(projectID))
	if err != nil {
		return apierrors.ErrGetSnippetYaml.InvalidParameter(err).ToResp(), nil
	}
	if project == nil {
		return apierrors.ErrGetSnippetYaml.InvalidParameter(fmt.Errorf("not find project: ID %v", projectIDString)).ToResp(), nil
	}
	apps, err := e.bdl.GetAppsByProject(project.ID, project.OrgID, project.Creator)
	if err != nil {
		return apierrors.ErrGetSnippetYaml.InvalidParameter(err).ToResp(), nil
	}

	appName := getAppNameFromYmlPath(ymlPath)
	var matchApp apistructs.ApplicationDTO
	for _, app := range apps.List {
		if app.Name == appName {
			matchApp = app
			break
		}
	}

	appID := matchApp.ID
	branch := getBranchFromYmlPath(ymlPath)

	if appID <= 0 {
		return apierrors.ErrGetSnippetYaml.InvalidParameter(errors.New("labels key appID value is empty")).ToResp(), nil
	}

	if branch == "" {
		return apierrors.ErrGetSnippetYaml.InvalidParameter(errors.New("labels key branch value is empty")).ToResp(), nil
	}

	app, err := e.bdl.GetApp(appID)
	if err != nil {
		return nil, apierrors.ErrGetApp.InternalError(err)
	}

	commit, err := e.bdl.GetGittarCommit(app.GitRepoAbbrev, branch)
	if err != nil {
		return nil, apierrors.ErrGetGittarCommit.InternalError(err)
	}

	repoAbbr := app.GitRepo
	commitID := commit.ID

	if repoAbbr == "" {
		return apierrors.ErrGetSnippetYaml.InvalidParameter(errors.New("repoAbbr value is empty")).ToResp(), nil
	}

	if commitID == "" {
		return apierrors.ErrGetSnippetYaml.InvalidParameter(errors.New("commitID value is empty")).ToResp(), nil
	}

	if strings.HasPrefix(req.Name, "/") {
		req.Name = strings.TrimPrefix(req.Name, "/")
	}

	yml, err := e.bdl.GetGittarFile(repoAbbr, commitID, req.Name, "", "")
	if err != nil {
		return apierrors.ErrGetSnippetYaml.InvalidParameter(err).ToResp(), nil
	}

	pipelineYml, err := pipelineyml.New([]byte(yml))
	if err != nil {
		return apierrors.ErrGetSnippetYaml.InvalidParameter(err).ToResp(), nil
	}

	workspace, err := e.fileTree.GetWorkspaceByBranch(strconv.Itoa(int(app.ProjectID)), branch)
	if err != nil {
		return apierrors.ErrGetSnippetYaml.InvalidParameter(err).ToResp(), nil
	}

	pipelineYml.Spec().LoopStagesActions(func(stage int, action *pipelineyml.Action) {
		// git-checkout 设置下 params
		if action.Type == "git-checkout" {
			if action.Params == nil {
				action.Params = map[string]interface{}{}
			}

			url := action.Params["uri"]
			paramsBranch := action.Params["branch"]

			if url == nil || url == "((gittar.repo))" {
				action.Params["uri"] = repoAbbr
			}
			if paramsBranch == nil || paramsBranch == "((gittar.branch))" {
				action.Params["branch"] = branch
			}
		}
		// 设置上额外的 env 给 pipeline 设置到 task 的 evn 上
		actionEnv := map[string]string{}
		actionEnv[apistructs.DiceApplicationId] = strconv.Itoa(int(appID))
		actionEnv[apistructs.DiceApplicationName] = app.Name
		actionEnv[apistructs.DiceWorkspaceEnv] = workspace
		if action.Type == "release" || action.Type == "dice" {
			actionEnv[apistructs.GittarBranchEnv] = branch
		}
		actionEnvByte, err := json.Marshal(actionEnv)
		if err != nil {
			return
		}

		// json格式化后端 env 设置到 snippetConfig 的一个标签上
		action.SnippetConfig = &pipelineyml.SnippetConfig{
			Labels: map[string]string{
				apistructs.LabelActionEnv: string(actionEnvByte),
			},
		}
	})
	spec, err := pipelineyml.GenerateYml(pipelineYml.Spec())
	if err != nil {
		return apierrors.ErrGetSnippetYaml.InvalidParameter(err).ToResp(), nil
	}

	return httpserver.OkResp(string(spec))
}

func getAppNameFromYmlPath(ymlPath string) string {
	ymlPath = strings.TrimPrefix(ymlPath, "/")
	return strings.SplitN(ymlPath, "/", 2)[0]
}

func getBranchFromYmlPath(ymlPath string) string {
	ymlPath = strings.TrimPrefix(ymlPath, "/")
	splits := strings.Split(ymlPath, "/")
	var branch string
	for i := 2; i < len(splits)-1; i++ {
		branch += splits[i] + "/"
	}
	branch = branch[:len(branch)-1]
	return branch
}
