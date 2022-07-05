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

	"github.com/sirupsen/logrus"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/services/apierrors"
	"github.com/erda-project/erda/internal/pkg/user"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
)

func (e *Endpoints) GetOrgNexus(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// get current user
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrGetOrgNexus.NotLogin().ToResp(), nil
	}

	// check org
	orgIDStr, ok := vars["orgID"]
	if !ok {
		return apierrors.ErrGetOrgNexus.MissingParameter("orgID").ToResp(), nil
	}
	orgResp, err := e.orgClient.GetOrg(apis.WithInternalClientContext(context.Background(), discover.SvcDOP),
		&orgpb.GetOrgRequest{IdOrName: orgIDStr})
	if err != nil {
		return apierrors.ErrGetOrgNexus.InvalidParameter(err).ToResp(), nil
	}
	org := orgResp.Data

	var req apistructs.OrgNexusGetRequest
	if r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return apierrors.ErrGetOrgNexus.InvalidParameter(err).ToResp(), nil
		}
	}

	// check permission
	if access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.OrgScope,
		ScopeID:  org.ID,
		Resource: apistructs.OrgResource,
		Action:   apistructs.GetAction,
	}); err != nil || !access.Access {
		if err != nil {
			logrus.Errorf("failed to check permission when get org nexus, err: %v", err)
		}
		return apierrors.ErrGetOrgNexus.AccessDenied().ToResp(), nil
	}

	nexusData, err := e.org.GetOrgLevelNexus(uint64(org.ID), &req)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(nexusData)
}

func (e *Endpoints) ShowOrgNexusPassword(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// get current user
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrGetOrgNexus.NotLogin().ToResp(), nil
	}

	var req apistructs.OrgNexusShowPasswordRequest
	if r.ContentLength == 0 {
		return apierrors.ErrShowOrgNexusPassword.MissingParameter("request body").ToResp(), nil
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrShowOrgNexusPassword.InvalidParameter(err).ToResp(), nil
	}

	// check permission
	if access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.OrgScope,
		ScopeID:  req.OrgID,
		Resource: apistructs.OrgResource,
		Action:   apistructs.GetAction,
	}); err != nil || !access.Access {
		if err != nil {
			logrus.Errorf("failed to check permission when get org nexus, err: %v", err)
		}
		return apierrors.ErrGetOrgNexus.AccessDenied().ToResp(), nil
	}

	data, err := e.org.ShowOrgNexusPassword(&req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(data)
}

func (e *Endpoints) GetNexusOrgDockerCredentialByImage(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// only internal invoke
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetNexusDockerCredentialByImage.InvalidParameter(err).ToResp(), nil
	}
	if !identityInfo.IsInternalClient() {
		return apierrors.ErrGetNexusDockerCredentialByImage.AccessDenied().ToResp(), nil
	}

	orgResp, err := e.orgClient.GetOrg(apis.WithInternalClientContext(context.Background(), discover.SvcDOP),
		&orgpb.GetOrgRequest{IdOrName: vars["orgID"]})
	if err != nil {
		return apierrors.ErrGetOrg.InvalidParameter(err).ToResp(), nil
	}
	org := orgResp.Data

	dockerPullUser, err := e.org.GetNexusOrgDockerCredential(uint64(org.ID), r.URL.Query().Get("image"))
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(dockerPullUser)
}
