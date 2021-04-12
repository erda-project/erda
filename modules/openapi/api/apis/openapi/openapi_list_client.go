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
	"bytes"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/auth"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/httpclient"
)

var OPENAPI_LIST_CLIENT = apis.ApiSpec{
	Path:   "/api/openapi/manager/clients",
	Scheme: "http",
	Method: "GET",
	Custom: func(rw http.ResponseWriter, req *http.Request) {
		token, err := auth.GetDiceClientToken()
		if err != nil {
			errStr := fmt.Sprintf("get token fail: %v", err)
			logrus.Error(errStr)
			http.Error(rw, errStr, http.StatusForbidden)
			return
		}
		logrus.Infof("diceclienttoken: %+v", token)
		var body bytes.Buffer
		r, err := httpclient.New(httpclient.WithCompleteRedirect()).Get(discover.UC()).Path("/api/open-client/manager/clients").
			Header("Authorization", "Bearer "+token.AccessToken).Do().Body(&body)
		if err != nil {
			errStr := fmt.Sprintf("list client fail: %v", err)
			logrus.Error(errStr)
			http.Error(rw, errStr, http.StatusForbidden)
			return
		}
		if !r.IsOK() {
			errStr := fmt.Sprintf("list client fail, statuscode: %d, body: %v", r.StatusCode(), body.String())
			logrus.Error(errStr)
			http.Error(rw, errStr, http.StatusForbidden)
			return
		}
		rw.Write([]byte(body.String()))
	},
	CheckLogin: false, // TODO:
	Doc: `
summary: 获取client列表
description: 认证： 通过 Authorization 头信息进行认证。 格式为“Bearer <token>”, 注意空格
produces:
  - application/json
`,
}
