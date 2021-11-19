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
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httputil"
)

// ApplicationsResources responses the resources list for every applications in the project
func (e *Endpoints) ApplicationsResources(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	l := logrus.WithField("func", "*Endpoints.ApplicationsResources")
	var (
		req apistructs.ApplicationsResourcesRequest
	)
	projectID, err := strconv.ParseUint(vars["projectID"], 10, 64)
	if err != nil {
		l.WithError(err).Errorln("failed to ParseUint projectID")
		return apierrors.ErrApplicationsResources.InvalidParameter(err).ToResp(), nil
	}

	orgIDStr := r.Header.Get(httputil.OrgHeader)
	orgID, err := strconv.ParseUint(orgIDStr, 10, 64)
	if err != nil {
		l.WithError(err).Errorln("failed to ParseUint orgID")
		return apierrors.ErrApplicationsResources.InvalidParameter(err).ToResp(), nil
	}

	userID, err := user.GetUserID(r)
	if err != nil {
		l.WithError(err).Errorln("failed to GetUserID")
		return apierrors.ErrApplicationsResources.InvalidParameter(err).ToResp(), nil
	}

	if err = r.ParseForm(); err != nil {
		l.WithError(err).Errorln("failed to ParseForm")
		return apierrors.ErrApplicationsResources.InvalidParameter(err).ToResp(), nil
	}

	req.ProjectID = projectID
	req.OrgID = orgID
	req.UserID = string(userID)
	req.Query = new(apistructs.ApplicationsResourceQuery)
	ParseApplicationsResourceQuery(req.Query, r.URL.Query())

	data, apiError := e.project.ApplicationsResources(ctx, &req)
	if apiError != nil {
		return apiError.ToResp(), nil
	}
	return httpserver.OkResp(data)
}

func ParseApplicationsResourceQuery(query *apistructs.ApplicationsResourceQuery, values url.Values) {
	if query == nil || len(values) == 0 {
		return
	}
	applicationIDs := values["applicationID"]
	for _, idStr := range applicationIDs {
		if id, err := strconv.ParseUint(idStr, 10, 64); err == nil {
			query.ApplicationsIDs = append(query.ApplicationsIDs, id)
		}
	}
	ownerIDs := values["ownerID"]
	for _, idStr := range ownerIDs {
		if id, err := strconv.ParseUint(idStr, 10, 64); err == nil {
			query.OwnerIDs = append(query.OwnerIDs, id)
		}
	}
	query.OrderBy = strings.Split(values.Get("orderBy"), ",")
	query.PageSize, _ = strconv.ParseUint(values.Get("pageSize"), 10, 64)
	query.PageNo, _ = strconv.ParseUint(values.Get("pageNo"), 10, 64)
	if query.PageSize == 0 {
		query.PageSize = 20
	}
	if query.PageNo == 0 {
		query.PageNo = 1
	}
}
