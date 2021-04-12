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

package apis

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi/conf"
)

var PRESALE_SURVEY = ApiSpec{
	Path:   "/api/survey/login",
	Scheme: "http",
	Method: "GET",
	Doc:    "summary: 售前调查 API（伪装成登录接口）",
	Custom: func(rw http.ResponseWriter, req *http.Request) {
		bdl := bundle.New(bundle.WithEventBox())
		f := req.URL.Query()
		content := ""
		content += "真实姓名　　: " + f.Get("realname") + "\n" +
			"手机号码　　: " + f.Get("mobile") + "\n" +
			"企业邮箱地址: " + f.Get("email") + "\n" +
			"所处职位　　: " + f.Get("position") + "\n" +
			"企业名称　　: " + f.Get("company") + "\n" +
			"企业规模　　: " + f.Get("company_size") + "\n" +
			"IT部门规模　: " + f.Get("it_size") + "\n" +
			"申请目的　　: " + f.Get("purpose") + "\n"
		msg := apistructs.MessageCreateRequest{
			Sender: "survey",
			Labels: map[apistructs.MessageLabel]interface{}{
				apistructs.DingdingLabel: []apistructs.Target{{Receiver: conf.SurveyDingding(), Secret: ""}},
			},
			Content: content,
		}
		if err := bdl.CreateMessage(&msg); err != nil {
			logrus.Warnf("failed to POST survey, err: %v", err)
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(500)
			rw.Write([]byte(`{"success":false,"err":{"msg":"失败"}}`))
			return
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(200)
		rw.Write([]byte(`{"success":true}`))
	},
}
