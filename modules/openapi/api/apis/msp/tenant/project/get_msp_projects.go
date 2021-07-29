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

package project

import (
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/orchestrator/utils"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

var MSP_PROJECT_LIST = apis.ApiSpec{
	Path:        "/api/msp/tenant/projects",
	BackendPath: "/api/msp/tenant/projects",
	Host:        "msp.marathon.l4lb.thisdcos.directory:8080",
	Scheme:      "http",
	Method:      "GET",
	CheckLogin:  true,
	CheckToken:  true,
	Custom:      attachProjectParams,
	IsOpenAPI:   true,
	Doc:         "GET MSP projects",
}

func attachProjectParams(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("User-ID")
	orgID := r.Header.Get("Org-ID")

	// get permissions
	client := httpclient.New()
	var perms apistructs.ScopeRoleList
	cr := client.Get(discover.CoreServices()).
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
	cr = client.Get(discover.MSP()).
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
