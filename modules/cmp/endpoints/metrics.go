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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/pkg/strutil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp/metrics"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

var permissionFailErr = fmt.Errorf("failed to get User-ID or Org-ID from request header")

// MetricsQuery handle query request
func (e *Endpoints) MetricsQuery(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		req apistructs.MetricsRequest
		err error
	)
	// get identity info
	i, resp := e.GetIdentity(r)
	if resp != nil {
		return httpserver.ErrResp(http.StatusInternalServerError, "InternalError", "identity not found")
	}
	// permission check
	err = e.PermissionCheck(i.UserID, i.OrgID, "", apistructs.GetAction)
	if err != nil {
		return httpserver.ErrResp(http.StatusInternalServerError, "InternalError", permissionFailErr.Error())
	}
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		logrus.Errorf("failed to unmarshal request: %v", err)
		return httpserver.ErrResp(http.StatusInternalServerError, "InternalError", "request json unmarshal err")
	}

	logrus.Infof("query metrics :%s %s %s names = %v ,ips = %v", req.ClusterName, req.ResourceKind, req.ResourceType, req.Names, req.IP)
	if strings.ToLower(req.ResourceKind) == metrics.Node {
		return e.metrics.QueryNodeResource(ctx, &req)
	} else {
		return e.metrics.QueryPodResource(ctx, &req)
	}
}

func ProxyMetrics(ctx context.Context, r *http.Request, vars map[string]string) error {

	// proxy
	r.URL.Scheme = "http"
	r.Host = discover.Monitor()
	r.URL.Host = discover.Monitor()
	r.URL.Path = strings.Replace(r.URL.Path, "/api/apim/metrics", "/api/metrics", 1)

	return nil
}

func InternalReverseHandler(handler func(context.Context, *http.Request, map[string]string) error) http.Handler {
	return &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			logrus.Debugf("start %s %s", r.Method, r.URL.String())

			handleRequest(r)

			err := handler(context.Background(), r, mux.Vars(r))
			if err != nil {
				logrus.Errorf("failed to handle request: %s (%v)", r.URL.String(), err)
				return
			}
		},
		FlushInterval: -1,
	}
}

func handleRequest(r *http.Request) {
	// base64 decode request body if declared in header
	if strutil.Equal(r.Header.Get(httpserver.Base64EncodedRequestBody), "true", true) {
		r.Body = ioutil.NopCloser(base64.NewDecoder(base64.StdEncoding, r.Body))
	}
}
