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

package endpoints

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/swagger"
	"github.com/erda-project/erda/pkg/swagger/oas3"
)

func (e *Endpoints) APIDocWebsocket(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return err
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		return err
	}

	apiErr := e.fileTreeSvc.Upgrade(w, r, &apistructs.WsAPIDocHandShakeReq{
		OrgID:    orgID,
		Identity: &identity,
		URIParams: &apistructs.FileTreeDetailURI{
			TreeName: "api-docs",
			Inode:    vars["inode"],
		},
	})
	if apiErr != nil {
		httpserver.WriteErr(w, apiErr.Code(), apiErr.Error())
	}

	return nil
}

func (e *Endpoints) ValidateSwagger(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return errorresp.New().InvalidParameter(err).ToResp(), nil
	}

	var response = map[string]interface{}{
		"success": true,
	}

	content, ok := body["content"]
	if !ok || content == "" {
		response["success"] = false
		response["err"] = "no content"
		return httpserver.OkResp(response)
	}

	v3, err := swagger.LoadFromData([]byte(content))
	if err != nil {
		response["success"] = false
		response["err"] = err.Error()
		return httpserver.OkResp(response)
	}

	if err = oas3.ValidateOAS3(context.TODO(), *v3); err != nil {
		response["success"] = false
		response["err"] = err.Error()
		return httpserver.OkResp(response)
	}

	return httpserver.OkResp(response)
}
