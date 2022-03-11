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
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dicehub/dbclient"
	"github.com/erda-project/erda/modules/dicehub/errcode"
	"github.com/erda-project/erda/modules/dicehub/response"
	"github.com/erda-project/erda/modules/dicehub/service/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

// GetDiceYAML GET /api/releases/<releaseId>/actions/get-dice 获取dice.yml内容处理
func (e *Endpoints) GetDiceYAML(w http.ResponseWriter, r *http.Request) {
	orgID, err := getPermissionHeader(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, errcode.HeaderMissing, "Header: [User-ID] or [Org-Id] is missing.")
		return
	}

	vars := mux.Vars(r)
	releaseID := vars["releaseId"]
	if releaseID == "" {
		logrus.Warn("Param [ReleaseID] is missing when get release info.")
		response.Error(w, http.StatusBadRequest, errcode.ParamMissing, "ReleaseID is missing.")
		return
	}

	logrus.Infof("getting dice.yml...releaseId: %s\n", releaseID)

	diceYAML, err := e.release.GetDiceYAML(orgID, releaseID)
	if err != nil {
		response.Error(w, http.StatusNotFound, errcode.ResourceNotFound, "release not found")
	}

	if strings.Contains(r.Header.Get("Accept"), "application/x-yaml") {
		w.Header().Set("Content-Type", "application/x-yaml; charset=utf-8")
		w.Write([]byte(diceYAML))
	} else { // 默认: application/json
		yaml, err := diceyml.New([]byte(diceYAML), false)
		if err != nil {
			logrus.Errorf("diceyml new error: %v", err)
			response.Error(w, http.StatusInternalServerError, errcode.InternalServerError, "Parse diceyml error.")
			return
		}
		diceJSON, err := yaml.JSON()
		if err != nil {
			logrus.Errorf("diceyml marshal error: %v", err)
			response.Error(w, http.StatusInternalServerError, errcode.InternalServerError, "Parse diceyml error.")
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(diceJSON))
	}
}

func getPermissionHeader(r *http.Request) (int64, error) {
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return 0, nil
	}
	return strconv.ParseInt(r.Header.Get("Org-ID"), 10, 64)
}

// DownloadRelease GET /api/releases/{releaseId}/actions/download 下载制品zip包
func (e *Endpoints) DownloadRelease(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	orgID, err := getPermissionHeader(r)
	if err != nil {
		return apierrors.ErrDownloadRelease.NotLogin()
	}

	releaseID := vars["releaseId"]
	if releaseID == "" {
		return apierrors.ErrDownloadRelease.MissingParameter("releaseId")
	}

	release, err := e.db.GetRelease(releaseID)
	if err != nil {
		return apierrors.ErrDownloadRelease.InternalError(err)
	}

	if !release.IsProjectRelease {
		return apierrors.ErrDownloadRelease.InvalidParameter("only project release can be downloaded")
	}
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrFormalRelease.NotLogin()
	}
	if !identityInfo.IsInternalClient() {
		hasAccess, err := e.hasReadAccess(identityInfo, release.ProjectID)
		if err != nil {
			return apierrors.ErrFormalRelease.InternalError(err)
		}
		if !hasAccess {
			return apierrors.ErrFormalRelease.AccessDenied()
		}
	}

	dir := fmt.Sprintf("%s_%s", release.ProjectName, release.Version)

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	releaseIDs, err := unmarshalApplicationReleaseList(release.ApplicationReleaseList)
	if err != nil {
		return apierrors.ErrDownloadRelease.InternalError(err)
	}

	appReleases, err := e.getAppReleases(releaseIDs)
	if err != nil {
		return apierrors.ErrDownloadRelease.InternalError(err)
	}

	for i := 0; i < len(appReleases); i++ {
		for j := 0; j < len(appReleases[i]); j++ {
			f, err := zw.Create(filepath.Join(dir, "dicefile", appReleases[i][j].ApplicationName, "dice.yml"))
			if err != nil {
				return apierrors.ErrDownloadRelease.InternalError(err)
			}
			if _, err := f.Write([]byte(appReleases[i][j].Dice)); err != nil {
				return apierrors.ErrDownloadRelease.InternalError(err)
			}
		}
	}

	u, err := e.bdl.GetCurrentUser(identityInfo.UserID)
	if err != nil {
		return apierrors.ErrDownloadRelease.InternalError(err)
	}
	org, err := e.bdl.GetOrg(orgID)
	if err != nil {
		return apierrors.ErrDownloadRelease.InternalError(err)
	}
	metadata, err := makeMetadata(org.DisplayName, u.Nick, release, appReleases)
	if err != nil {
		return apierrors.ErrDownloadRelease.InternalError(err)
	}

	f, err := zw.Create(filepath.Join(dir, "metadata.yml"))
	if err != nil {
		return apierrors.ErrDownloadRelease.InternalError(err)
	}
	if _, err := f.Write(metadata); err != nil {
		return apierrors.ErrDownloadRelease.InternalError(err)
	}

	if err := zw.Close(); err != nil {
		return apierrors.ErrDownloadRelease.InternalError(err)
	}

	w.Header().Add("Content-type", "application/zip")
	w.Header().Add("Content-Disposition", "attachment;fileName="+dir+".zip")

	if _, err := io.Copy(w, buf); err != nil {
		return apierrors.ErrDownloadRelease.InternalError(err)
	}
	return nil
}

func (e *Endpoints) getAppReleases(releaseList [][]string) ([][]dbclient.Release, error) {
	var list []string
	for i := 0; i < len(releaseList); i++ {
		list = append(list, releaseList[i]...)
	}

	releases, err := e.db.GetReleases(list)
	if err != nil {
		return nil, err
	}

	id2Release := make(map[string]*dbclient.Release)
	for i := 0; i < len(releases); i++ {
		id2Release[releases[i].ReleaseID] = &releases[i]
	}

	appReleases := make([][]dbclient.Release, len(releaseList))
	for i := 0; i < len(releaseList); i++ {
		appReleases[i] = make([]dbclient.Release, len(releaseList[i]))
		for j := 0; j < len(releaseList[i]); j++ {
			appReleases[i][j] = *id2Release[releaseList[i][j]]
		}
	}
	return appReleases, nil
}

// hasReadAccess check whether user has access to get project
func (e *Endpoints) hasReadAccess(identityInfo apistructs.IdentityInfo, projectID int64) (bool, error) {
	access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   identityInfo.UserID,
		Scope:    apistructs.ProjectScope,
		ScopeID:  uint64(projectID),
		Resource: apistructs.ProjectResource,
		Action:   apistructs.GetAction,
	})
	if err != nil {
		return false, err
	}
	if !access.Access {
		return false, nil
	}
	return true, nil
}

// hasWriteAccess check whether user is project owner or project lead
func (e *Endpoints) hasWriteAccess(identity apistructs.IdentityInfo, projectID int64, isProjectRelease bool, applicationID int64) (bool, error) {
	req := &apistructs.ScopeRoleAccessRequest{
		Scope: apistructs.Scope{
			Type: apistructs.ProjectScope,
			ID:   strconv.FormatInt(projectID, 10),
		},
	}
	rsp, err := e.bdl.ScopeRoleAccess(identity.UserID, req)
	if err != nil {
		return false, err
	}

	hasProjectAccess := false
	for _, role := range rsp.Roles {
		if role == bundle.RoleProjectOwner || role == bundle.RoleProjectLead || role == bundle.RoleProjectPM {
			hasProjectAccess = true
			break
		}
	}

	if isProjectRelease || hasProjectAccess {
		return hasProjectAccess, nil
	}

	req = &apistructs.ScopeRoleAccessRequest{
		Scope: apistructs.Scope{
			Type: apistructs.AppScope,
			ID:   strconv.FormatInt(applicationID, 10),
		},
	}
	rsp, err = e.bdl.ScopeRoleAccess(identity.UserID, req)
	if err != nil {
		logrus.Errorf("failed to check app access for release of app %d, %v", applicationID, err)
		return hasProjectAccess, nil
	}

	hasAppAccess := false
	for _, role := range rsp.Roles {
		if role == bundle.RoleAppOwner || role == bundle.RoleAppLead {
			hasAppAccess = true
			break
		}
	}
	return hasAppAccess, nil
}

func unmarshalApplicationReleaseList(str string) ([][]string, error) {
	var list [][]string
	if err := json.Unmarshal([]byte(str), &list); err != nil {
		return nil, err
	}
	return list, nil
}

func makeMetadata(orgName, userName string, release *dbclient.Release, appReleases [][]dbclient.Release) ([]byte, error) {
	appList := make([][]apistructs.AppMetadata, len(appReleases))
	for i := 0; i < len(appReleases); i++ {
		appList[i] = make([]apistructs.AppMetadata, len(appReleases[i]))
		for j := 0; j < len(appReleases[i]); j++ {
			labels := make(map[string]string)
			if err := json.Unmarshal([]byte(appReleases[i][j].Labels), &labels); err != nil {
				logrus.Errorf("failed to unmarshal labels for release %s, %v", appReleases[i][j].ReleaseID, err)
			}
			appList[i][j] = apistructs.AppMetadata{
				AppName:          appReleases[i][j].ApplicationName,
				GitBranch:        labels["gitBranch"],
				GitCommitID:      labels["gitCommitId"],
				GitCommitMessage: labels["gitCommitMessage"],
				GitRepo:          labels["gitRepo"],
				ChangeLog:        appReleases[i][j].Changelog,
				Version:          appReleases[i][j].Version,
			}
		}
	}
	releaseMeta := apistructs.ReleaseMetadata{
		Org:       orgName,
		Source:    "erda",
		Author:    userName,
		Version:   release.Version,
		Desc:      release.Desc,
		ChangeLog: release.Changelog,
		AppList:   appList,
	}
	return yaml.Marshal(releaseMeta)
}

func getRealIP(request *http.Request) string {
	ra := request.RemoteAddr
	if ip := request.Header.Get("X-Forwarded-For"); ip != "" {
		ra = strings.Split(ip, ", ")[0]
	} else if ip := request.Header.Get("X-Real-IP"); ip != "" {
		ra = ip
	} else {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
}
