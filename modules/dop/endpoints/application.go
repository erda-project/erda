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

package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/dop/types"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateApplication 创建应用
func (e *Endpoints) CreateApplication(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateApplication.InvalidParameter(err).ToResp(), nil
	}

	if r.Body == nil {
		return apierrors.ErrCreateApplication.MissingParameter("body is nil").ToResp(), nil
	}

	var applicationCreateReq apistructs.ApplicationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&applicationCreateReq); err != nil {
		return apierrors.ErrCreateApplication.InvalidParameter("can't decode body").ToResp(), nil
	}
	if !strutil.IsValidPrjOrAppName(applicationCreateReq.Name) {
		return apierrors.ErrCreateApplication.InvalidParameter(errors.Errorf("app name is invalid %s",
			applicationCreateReq.Name)).ToResp(), nil
	}
	logrus.Infof("request body: %+v", applicationCreateReq)

	// check param
	if err := checkApplicationCreateParam(applicationCreateReq); err != nil {
		return apierrors.ErrCreateApplication.InvalidParameter(err).ToResp(), nil
	}

	// check permission
	req := apistructs.PermissionCheckRequest{
		UserID:   identity.UserID,
		Scope:    apistructs.ProjectScope,
		ScopeID:  applicationCreateReq.ProjectID,
		Resource: apistructs.AppResource,
		Action:   apistructs.CreateAction,
	}
	if access, err := e.bdl.CheckPermission(&req); err != nil || !access.Access {
		return apierrors.ErrCreateApplication.AccessDenied().ToResp(), nil
	}

	// create app
	applicationDTO, err := e.bdl.CreateApp(applicationCreateReq, identity.UserID)
	if err != nil {
		return apierrors.ErrCreateApplication.InternalError(err).ToResp(), nil
	}

	// generate extra info
	e.generateExtraInfo(int64(applicationDTO.ID), int64(applicationDTO.ProjectID))

	return httpserver.OkResp(applicationDTO)
}

func checkApplicationCreateParam(applicationCreateReq apistructs.ApplicationCreateRequest) error {
	if applicationCreateReq.Name == "" {
		return errors.Errorf("invalid request, name is empty")
	}
	if applicationCreateReq.ProjectID == 0 {
		return errors.Errorf("invalid request, projectId is empty")
	}
	err := applicationCreateReq.Mode.CheckAppMode()
	return err
}

// generateExtraInfo 创建应用时，自动生成extra信息
func (e *Endpoints) generateExtraInfo(applicationID, projectID int64) string {
	// 初始化DEV、TEST、STAGING、PROD四个环境namespace，eg: "DEV.configNamespace":"app-107-DEV"
	workspaces := []apistructs.DiceWorkspace{
		types.DefaultWorkspace,
		types.DevWorkspace,
		types.TestWorkspace,
		types.StagingWorkspace,
		types.ProdWorkspace,
	}

	relatedNamespaces := make([]string, 0, len(workspaces)-1)
	var defaultNamespace string
	extra := make(map[string]string, len(workspaces))
	for _, v := range workspaces {
		key := strutil.Concat(string(v), ".configNamespace")
		value := strutil.Concat("app-", strconv.FormatInt(applicationID, 10), "-", string(v))
		extra[key] = value

		if v == types.DefaultWorkspace {
			defaultNamespace = value
		} else {
			relatedNamespaces = append(relatedNamespaces, value)
		}
	}

	// 创建 DEV/TEST/STAGING/PROD/DEFAULT namespace
	for _, v := range extra {
		namespaceCreateReq := apistructs.NamespaceCreateRequest{
			ProjectID: projectID,
			Dynamic:   true,
			Name:      v,
			IsDefault: strings.Contains(v, string(types.DefaultWorkspace)),
		}
		// 创建namespace
		_, err := e.namespace.Create(&namespaceCreateReq)
		if err != nil {
			logrus.Errorf(err.Error())
		}
	}

	// 创建 namespace relations
	relationCreateReq := apistructs.NamespaceRelationCreateRequest{
		RelatedNamespaces: relatedNamespaces,
		DefaultNamespace:  defaultNamespace,
	}

	if err := e.namespace.CreateRelation(&relationCreateReq); err != nil {
		logrus.Errorf(err.Error())
	}

	extraInfo, err := json.Marshal(extra)
	if err != nil {
		logrus.Errorf("failed to marshal extra info, (%v)", err)
	}
	return string(extraInfo)
}

// getWorkspaces return workspaces
func (e *Endpoints) getWorkspaces(extraStr string, projectID uint64) []apistructs.ApplicationWorkspace {
	var extra map[string]string
	if err := json.Unmarshal([]byte(extraStr), &extra); err != nil {
		extra = make(map[string]string)
	}
	project, err := e.bdl.GetProject(projectID)
	if err != nil {
		logrus.Error(err)
	}

	workspaces := make([]apistructs.ApplicationWorkspace, 0, len(extra))
	for k, v := range extra {
		if strings.Contains(k, ".") {
			env := strings.Split(k, ".")[0]
			workspace := apistructs.ApplicationWorkspace{
				Workspace:       env,
				ConfigNamespace: v,
			}
			if project != nil {
				workspace.ClusterName = project.ClusterConfig[env]
			}
			workspaces = append(workspaces, workspace)
		}
	}
	return workspaces
}

// DeleteApplication delete application
func (e *Endpoints) DeleteApplication(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	// 获取当前用户
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDeleteApplication.NotLogin().ToResp(), nil
	}

	// 检查applicationID合法性
	applicationID, err := strutil.Atoi64(vars["applicationID"])
	if err != nil {
		return apierrors.ErrDeleteApplication.InvalidParameter(err).ToResp(), nil
	}

	// 操作鉴权
	req := apistructs.PermissionCheckRequest{
		UserID:   identity.UserID,
		Scope:    apistructs.AppScope,
		ScopeID:  uint64(applicationID),
		Resource: apistructs.AppResource,
		Action:   apistructs.DeleteAction,
	}
	if access, err := e.bdl.CheckPermission(&req); err != nil || !access.Access {
		return apierrors.ErrDeleteApplication.AccessDenied().ToResp(), nil
	}
	appDto, err := e.bdl.DeleteApp(uint64(applicationID), identity.UserID)
	if err != nil {
		return apierrors.ErrDeleteApplication.InternalError(err).ToResp(), nil
	}

	// delete extra
	e.deleteExtraInfo(appDto.Extra, identity)

	// delete issue
	if err = e.db.DeleteIssueAppRelationsByApp(applicationID); err != nil {
		logrus.Warnf("failed to delete mr relations, %v", err)
	}

	// delete branch rule
	if err = e.db.DeleteBranchRuleByScope(apistructs.AppScope, applicationID); err != nil {
		logrus.Warnf("failed to delete app branch rules, (%v)", err)
	}

	return httpserver.OkResp(appDto)
}

// 从addon platform删除namespace
func (e *Endpoints) deleteExtraInfo(extra string, identityInfo apistructs.IdentityInfo) {
	var extraMap map[string]string
	if err := json.Unmarshal([]byte(extra), &extraMap); err != nil {
		logrus.Warnf("failed to unmarshal extra, (%v)", err)
		return
	}
	for _, v := range extraMap {
		// 删除 namespace relation(删除namespace relation须在删除namespace前进行，后续须推动优化)
		if strings.Contains(v, string(types.DefaultWorkspace)) {
			if err := e.namespace.DeleteRelation(nil, nil, v, identityInfo.UserID); err != nil {
				logrus.Warnf(err.Error())
			}
		}
		// 删除 namespace
		if err := e.namespace.DeleteNamespace(nil, v, identityInfo); err != nil {
			logrus.Warnf(err.Error())
		}
	}
}
