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

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httputil"
)

// GetResourceApplicationTrend is the echarts api for applications resource trend
func (e *Endpoints) GetResourceApplicationTrend(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var l = logrus.WithField("func", "*Endpoints.GetResourceApplicationTrend")
	if err := r.ParseForm(); err != nil {
		l.WithError(err).Errorln("failed to ParseForm")
		return httpserver.ErrResp(http.StatusBadRequest, "", "invalid query params")
	}
	var (
		req = new(apistructs.GetResourceApplicationTrendReq)
		q   = r.URL.Query()
	)
	req.OrgID = r.Header.Get(httputil.OrgHeader)
	req.UserID = r.Header.Get(httputil.UserHeader)
	req.ProjectID = vars["projectID"]
	req.Query = new(apistructs.GetResourceApplicationTrendReqQuery)
	req.Query.Start = q.Get("start")
	req.Query.End = q.Get("end")
	req.Query.Interval = q.Get("interval")
	req.Query.ResourceType = q.Get("resourceType")
	if err := req.Validate(); err != nil {
		return httpserver.ErrResp(http.StatusBadRequest, "", err.Error())
	}
	data, err := e.project.GetApplicationTrend(ctx, req)
	if err != nil {
		return httpserver.ErrResp(http.StatusInternalServerError, "", err.Error())
	}
	return httpserver.OkResp(data)
}
