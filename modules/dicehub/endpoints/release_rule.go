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

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dicehub/service/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
)

// CreateRule creates release rule record
func (e *Endpoints) CreateRule(ctx context.Context, r *http.Request, _ map[string]string) (httpserver.Responser, error) {
	request := ctx.Value("CreateUpdateDeleteReleaseRuleRequest").(*apistructs.CreateUpdateDeleteReleaseRuleRequest)
	data, apiError := e.releaseRule.Create(request)
	if apiError != nil {
		return apiError.ToResp(), nil
	}
	return httpserver.OkResp(data)
}

// ListRules lists release rules for the given project
func (e *Endpoints) ListRules(ctx context.Context, r *http.Request, _ map[string]string) (httpserver.Responser, error) {
	request := ctx.Value("CreateUpdateDeleteReleaseRuleRequest").(*apistructs.CreateUpdateDeleteReleaseRuleRequest)
	data, apiError := e.releaseRule.List(request)
	if apiError != nil {
		return apiError.ToResp(), nil
	}
	return httpserver.OkResp(data)
}

// UpdateRule updates the given release rule
func (e *Endpoints) UpdateRule(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	request := ctx.Value("CreateUpdateDeleteReleaseRuleRequest").(*apistructs.CreateUpdateDeleteReleaseRuleRequest)
	data, apiError := e.releaseRule.Update(request)
	if apiError != nil {
		return apiError.ToResp(), nil
	}
	return httpserver.OkResp(data)
}

// DeleteRule deletes the given release rule
func (e *Endpoints) DeleteRule(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	request := ctx.Value("CreateUpdateDeleteReleaseRuleRequest").(*apistructs.CreateUpdateDeleteReleaseRuleRequest)
	apiError := e.releaseRule.Delete(request)
	if apiError != nil {
		return apiError.ToResp(), nil
	}
	return httpserver.OkResp(nil)
}

func (e *Endpoints) ReleaseRuleAuth(handler httpserver.Handler) httpserver.Handler {
	return func(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
		if apiError := e.releaseRuleAuth(&ctx, r, vars); apiError != nil {
			return apiError.ToResp(), nil
		}
		return handler(ctx, r, vars)
	}
}

func (e *Endpoints) releaseRuleAuth(ctx *context.Context, r *http.Request, vars map[string]string) *errorresp.APIError {
	var l = logrus.WithField("func", "*Endpoints.ReleaseRuleAuth")

	// 常规性地参数检查
	userIDStr, err := user.GetUserID(r)
	if err != nil {
		l.WithError(err).Errorln("failed to GetUserID")
		return apierrors.ErrAuthReleaseRule.NotLogin()

	}
	userID, err := strconv.ParseUint(userIDStr.String(), 10, 32)
	if err != nil {
		l.WithError(err).WithField("userID", userIDStr).Errorln("failed to ParseUint")
		return apierrors.ErrAuthReleaseRule.InvalidParameter("invalid userID").NotLogin()
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		l.WithError(err).Errorln("failed to GetOrgID")
		return apierrors.ErrAuthReleaseRule.NotLogin()
	}
	projectIDStr := r.URL.Query().Get("projectID")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		l.WithError(err).WithField("projectID", projectIDStr).Errorln("failed to parseUint")
		return apierrors.ErrAuthReleaseRule.InvalidParameter("invalid query parameter projectID")
	}

	// 鉴权
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodDelete:
		identity, err := user.GetIdentityInfo(r)
		if err != nil {
			l.WithError(err).Errorln("failed to GetIdentityInfo")
			return apierrors.ErrAuthReleaseRule.NotLogin()
		}
		access, err := e.hasWriteAccess(identity, int64(projectID))
		if err != nil {
			l.WithError(err).Errorln("failed to hasWriteAccess")
			return apierrors.ErrAuthReleaseRule.InternalError(err)
		}
		if !access {
			return apierrors.ErrAuthReleaseRule.AccessDenied()
		}
	}

	// 封装参数
	var request = apistructs.CreateUpdateDeleteReleaseRuleRequest{
		OrgID:     orgID,
		ProjectID: projectID,
		UserID:    userID,
		RuleID:    vars["id"],
		Body:      new(apistructs.CreateUpdateReleaseRuleRequestBody),
	}
	switch r.Method {
	case http.MethodPost, http.MethodPut:
		if err := json.NewDecoder(r.Body).Decode(request.Body); err != nil {
			return apierrors.ErrAuthReleaseRule.InvalidParameter("can not read request body")
		}
	}
	*ctx = context.WithValue(*ctx, "CreateUpdateDeleteReleaseRuleRequest", &request)
	return nil
}
