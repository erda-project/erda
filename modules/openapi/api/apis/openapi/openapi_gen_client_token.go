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

package openapi

import (
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/ucauth"
)

var OPENAPI_GEN_CLIENT_TOKEN = apis.ApiSpec{
	Path:       "/api/openapi/client-token",
	Scheme:     "http",
	Method:     "POST",
	CheckLogin: false,
	Custom: func(rw http.ResponseWriter, req *http.Request) {
		basic := req.Header.Get("Authorization")
		if basic == "" {
			http.Error(rw, "not provide Basic Authorization header", http.StatusBadRequest)
			return
		}

		oauthToken, err := ucauth.GenClientToken(discover.UC(), basic)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		res, err := json.Marshal(oauthToken)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusForbidden)
			return
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.Write(res)
	},
	Doc: `
summary: client token 发放接口
description: 通过 header Basic 认证. Basic Header："Basic " + base64(<clientid>+":"+<clientsecret>)
produces:
  - application/json
responses:
  '200':
    description: OK
    schema:
      type: object
      properties:
        access_token:
          type: string
        token_type:
          type: string
        refresh_token:
          type: string
        expires_in:
          type: int64
        scope:
          type: string
        jti:
          type: string

  '400':
    description: 没有提供 Authorization header`,
}
