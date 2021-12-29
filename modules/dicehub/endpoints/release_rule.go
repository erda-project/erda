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
)

// CreateRule creates release rule record
func (e *Endpoints) CreateRule(ctx context.Context, r *http.Request, _ map[string]string) (httpserver.Responser, error) {
	var l = logrus.WithField("func", "*Endpoint.CreateRule")
	userIDStr, err := user.GetUserID(r)
	if err != nil {
		l.WithError(err).Errorln("failed to GetUserID")
		return apierrors.ErrCreateReleaseRule.NotLogin().ToResp(), nil
	}
	userID, err := strconv.ParseUint(userIDStr.String(), 10, 32)
	if err != nil {
		l.WithError(err).WithField("userID", userIDStr).Errorln("failed to ParseUint")
		return apierrors.ErrCreateReleaseRule.InvalidParameter("invalid userID").NotLogin().ToResp(), nil
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		l.WithError(err).Errorln("failed to GetOrgID")
		return apierrors.ErrListReleaseRule.NotLogin().ToResp(), nil
	}
	projectIDStr := r.URL.Query().Get("projectID")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		l.WithError(err).WithField("projectID", projectIDStr).Errorln("failed to parseUint")
		return apierrors.ErrCreateReleaseRule.InvalidParameter("invalid query parameter projectID").ToResp(), nil
	}
	var request = apistructs.CreateUpdateDeleteReleaseRuleRequest{
		OrgID:     orgID,
		ProjectID: projectID,
		UserID:    userID,
		RuleID:    0,
		Body:      new(apistructs.CreateUpdateReleaseRuleRequestBody),
	}
	if err = json.NewDecoder(r.Body).Decode(request.Body); err != nil {
		l.WithError(err).Errorln("failed to Decode request.Body")
		return apierrors.ErrCreateReleaseRule.InvalidParameter("failed to Decode request.Body").ToResp(), nil
	}
	data, apiError := e.releaseRule.Create(&request)
	if apiError != nil {
		return apiError.ToResp(), nil
	}
	return httpserver.OkResp(data)
}

// ListRules lists release rules for the given project
func (e *Endpoints) ListRules(ctx context.Context, r *http.Request, _ map[string]string) (httpserver.Responser, error) {
	var l = logrus.WithField("func", "*Endpoint.CreateRule")
	userIDStr, err := user.GetUserID(r)
	if err != nil {
		l.WithError(err).Errorln("failed to GetUserID")
		return apierrors.ErrListReleaseRule.NotLogin().ToResp(), nil
	}
	userID, err := strconv.ParseUint(userIDStr.String(), 10, 32)
	if err != nil {
		l.WithError(err).WithField("userID", userIDStr).Errorln("failed to ParseUint")
		return apierrors.ErrListReleaseRule.InvalidParameter("invalid userID").NotLogin().ToResp(), nil
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		l.WithError(err).Errorln("failed to GetOrgID")
		return apierrors.ErrListReleaseRule.NotLogin().ToResp(), nil
	}
	projectIDStr := r.URL.Query().Get("projectID")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		l.WithError(err).WithField("projectID", projectIDStr).Errorln("failed to parseUint")
		return apierrors.ErrListReleaseRule.InvalidParameter("invalid query parameter projectID").ToResp(), nil
	}
	var request = apistructs.ListReleaseRuleRequest{
		OrgID:     orgID,
		ProjectID: projectID,
		UserID:    userID,
	}
	data, apiError := e.releaseRule.List(&request)
	if apiError != nil {
		return apiError.ToResp(), nil
	}
	return httpserver.OkResp(data)
}

// UpdateRule updates the given release rule
func (e *Endpoints) UpdateRule(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var l = logrus.WithField("func", "*Endpoint.UpdateRule")
	userIDStr, err := user.GetUserID(r)
	if err != nil {
		l.WithError(err).Errorln("failed to GetUserID")
		return apierrors.ErrUpdateReleaseRule.NotLogin().ToResp(), nil
	}
	userID, err := strconv.ParseUint(userIDStr.String(), 10, 32)
	if err != nil {
		l.WithError(err).WithField("userID", userIDStr).Errorln("failed to ParseUint")
		return apierrors.ErrUpdateReleaseRule.InvalidParameter("invalid userID").NotLogin().ToResp(), nil
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		l.WithError(err).Errorln("failed to GetOrgID")
		return apierrors.ErrUpdateReleaseRule.NotLogin().ToResp(), nil
	}
	projectIDStr := r.URL.Query().Get("projectID")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		l.WithError(err).WithField("projectID", projectIDStr).Errorln("failed to parseUint")
		return apierrors.ErrUpdateReleaseRule.InvalidParameter("invalid query parameter projectID").ToResp(), nil
	}
	idStr := vars["id"]
	ruleID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		l.WithError(err).WithField("id", idStr).Errorln("failed to ParseUint")
		return apierrors.ErrUpdateReleaseRule.InvalidParameter("invalid rule id").ToResp(), nil
	}
	var request = apistructs.CreateUpdateDeleteReleaseRuleRequest{
		OrgID:     orgID,
		ProjectID: projectID,
		UserID:    userID,
		RuleID:    ruleID,
		Body:      new(apistructs.CreateUpdateReleaseRuleRequestBody),
	}
	data, apiError := e.releaseRule.Update(&request)
	if apiError != nil {
		return apiError.ToResp(), nil
	}
	return httpserver.OkResp(data)
}

// DeleteRule deletes the given release rule
func (e *Endpoints) DeleteRule(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var l = logrus.WithField("func", "*Endpoint.DeleteRule")
	userIDStr, err := user.GetUserID(r)
	if err != nil {
		l.WithError(err).Errorln("failed to GetUserID")
		return apierrors.ErrDeleteReleaseRule.NotLogin().ToResp(), nil
	}
	userID, err := strconv.ParseUint(userIDStr.String(), 10, 32)
	if err != nil {
		l.WithError(err).WithField("userID", userIDStr).Errorln("failed to ParseUint")
		return apierrors.ErrDeleteReleaseRule.InvalidParameter("invalid userID").NotLogin().ToResp(), nil
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		l.WithError(err).Errorln("failed to GetOrgID")
		return apierrors.ErrDeleteReleaseRule.NotLogin().ToResp(), nil
	}
	projectIDStr := r.URL.Query().Get("projectID")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		l.WithError(err).WithField("projectID", projectIDStr).Errorln("failed to parseUint")
		return apierrors.ErrDeleteReleaseRule.InvalidParameter("invalid query parameter projectID").ToResp(), nil
	}
	idStr := vars["id"]
	ruleID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		l.WithError(err).WithField("id", idStr).Errorln("failed to ParseUint")
		return apierrors.ErrDeleteReleaseRule.InvalidParameter("invalid rule id").ToResp(), nil
	}
	var request = apistructs.CreateUpdateDeleteReleaseRuleRequest{
		OrgID:     orgID,
		ProjectID: projectID,
		UserID:    userID,
		RuleID:    ruleID,
		Body:      nil,
	}
	apiError := e.releaseRule.Delete(&request)
	if apiError != nil {
		return apiError.ToResp(), nil
	}
	return httpserver.OkResp(nil)
}
