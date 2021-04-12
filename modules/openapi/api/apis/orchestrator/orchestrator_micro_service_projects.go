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

package orchestrator

import (
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/conf"
	"github.com/erda-project/erda/modules/orchestrator/utils"
	"github.com/erda-project/erda/modules/pkg/innerdomain"
	"github.com/erda-project/erda/pkg/httpclient"
)

var ORCHESTRATOR_MICRO_SERVICE_PROJECTS = apis.ApiSpec{
	Path:         "/api/microservice/projects",
	BackendPath:  "/api/microservice/projects",
	Host:         "orchestrator.marathon.l4lb.thisdcos.directory:8081",
	Scheme:       "http",
	Method:       "GET",
	ResponseType: apistructs.MicroServiceProjectResponse{},
	Custom:       attachProjectParams,
	CheckLogin:   true,
	CheckToken:   true,
	IsOpenAPI:    true,
	Doc:          `summary: 微服务项目列表`,
	Group:        "addons",
}

func attachProjectParams(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("User-ID")
	orgID := r.Header.Get("Org-ID")

	// get permissions
	client := httpclient.New()
	var perms apistructs.ScopeRoleList
	cmdb := innerdomain.MustParseWithEnv("cmdb.marathon.l4lb.thisdcos.directory", conf.UseK8S())
	cr := client.Get(cmdb+":9093").
		Path("/api/permissions").
		Header("User-ID", userID).
		Header("Org-ID", orgID)
	if err := utils.DoJson(cr, &perms); err != nil {
		ErrFromError(w, err)
	}

	params := r.URL.Query()
	params.Del("projectId")
	for _, p := range perms.List {
		if p.Scope.Type == apistructs.ProjectScope && p.Access {
			params.Add("projectId", p.Scope.ID)
		}
	}

	var data json.RawMessage
	orchestrator := innerdomain.MustParseWithEnv("orchestrator.marathon.l4lb.thisdcos.directory", conf.UseK8S())
	cr = client.Get(orchestrator+":8081").
		Header("User-ID", userID).
		Header("Org-ID", orgID).
		Path(r.URL.Path).Params(params)

	if err := utils.DoJson(cr, &data); err != nil {
		ErrFromError(w, err)
		return
	}

	Succ(w, data)
}

func Succ(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	b, _ := json.Marshal(utils.Resp{
		Success: true,
		Data:    data,
	})
	w.Write(b)
}
func ErrFromError(w http.ResponseWriter, error error) {
	Err(w, &apistructs.ErrorResponse{Code: "InternalError", Msg: error.Error()}, http.StatusInternalServerError)
}
func Err(w http.ResponseWriter, err *apistructs.ErrorResponse, httpCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpCode)
	b, _ := json.Marshal(utils.Resp{
		Success: false,
		Err:     err,
	})
	w.Write(b)
}
