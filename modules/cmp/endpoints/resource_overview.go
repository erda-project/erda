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
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httputil"
)

func (e *Endpoints) ResourceOverviewReport(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	logrus.Debugln("ResourceOverviewReport")

	orgIDStr := r.Header.Get(httputil.OrgHeader)
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil {
		logrus.WithError(err).Errorln("failed to parse orgID")
		return httpserver.ErrResp(0, "", err.Error()) // todo:
	}
	if err := r.ParseForm(); err != nil {
		logrus.WithError(err).Errorln("failed to ParseForm")
		return httpserver.ErrResp(0, "", err.Error()) // todo:
	}

	value := r.URL.Query()
	clusterNames := value["clusterName"]
	cpuPerNodeStr := value.Get("cpuPerNode")
	memPerNodeStr := value.Get("memPerNode")
	cpuPerNode, err := strconv.ParseUint(cpuPerNodeStr, 10, 64)
	if err != nil {
		cpuPerNode = 8
	}
	memPerNode, err := strconv.ParseUint(memPerNodeStr, 10, 64)
	if err != nil {
		memPerNode = 32
	}
	logrus.Debugf("params: orgID: %v, clusterNames: %v, cpuPerNode: %v, memPerNode: %v", orgID, clusterNames, cpuPerNode, memPerNode)

	report, err := e.reportTable.GetResourceOverviewReport(ctx, orgID, clusterNames, cpuPerNode, memPerNode)
	if err != nil {
		logrus.WithError(err).Errorln("failed to GetResourceOverviewReport")
		return httpserver.ErrResp(0, "", err.Error())
	}

	return httpserver.OkResp(report)
}
