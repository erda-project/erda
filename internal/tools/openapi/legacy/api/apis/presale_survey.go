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

package apis

import (
	"fmt"
	"net/http"

	"github.com/gorilla/schema"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/tools/openapi/legacy/conf"
)

type detail struct {
	RealName    string `schema:"realname,required"`
	Mobile      string `schema:"mobile,required"`
	Email       string `schema:"email,required"`
	Position    string `schema:"position,required"`
	Company     string `schema:"company,required"`
	CompanySize string `schema:"company_size,required"`
	ITSize      string `schema:"it_size,required"`
	Purpose     string `schema:"purpose,required"`
}

var queryDecoder = schema.NewDecoder()

func validatePurveyDetail(d *detail, m map[string][]string) error {
	return queryDecoder.Decode(d, m)
}

var PRESALE_SURVEY = ApiSpec{
	Path:   "/api/survey/login",
	Scheme: "http",
	Method: "GET",
	Doc:    "summary: 售前调查 API（伪装成登录接口）",
	Custom: func(rw http.ResponseWriter, req *http.Request) {
		var d detail
		if err := validatePurveyDetail(&d, req.URL.Query()); err != nil {
			logrus.Errorf("failed to parse survey detail, err: %v", err)
			errorResp(rw)
			return
		}
		msg := apistructs.MessageCreateRequest{
			Sender: "survey",
			Labels: map[apistructs.MessageLabel]interface{}{
				apistructs.DingdingLabel: []apistructs.Target{{Receiver: conf.SurveyDingding(), Secret: ""}},
			},
			Content: makeDingdingMessage(d),
		}
		bdl := bundle.New(bundle.WithErdaServer())
		if err := bdl.CreateMessage(&msg); err != nil {
			logrus.Warnf("failed to POST survey, err: %v", err)
			errorResp(rw)
			return
		}
		successResp(rw)
	},
}

func makeDingdingMessage(d detail) string {
	content := ""
	content += fmt.Sprintf("真实姓名　　: %s\n", d.RealName)
	content += fmt.Sprintf("手机号码　　: %s\n", d.Mobile)
	content += fmt.Sprintf("企业邮箱地址: %s\n", d.Email)
	content += fmt.Sprintf("所处职位　　: %s\n", d.Position)
	content += fmt.Sprintf("企业名称　　: %s\n", d.Company)
	content += fmt.Sprintf("企业规模　　: %s\n", d.CompanySize)
	content += fmt.Sprintf("IT部门规模　: %s\n", d.ITSize)
	content += fmt.Sprintf("申请目的　　: %s\n", d.Purpose)
	return content
}
func errorResp(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(500)
	rw.Write([]byte(`{"success":false,"err":{"msg":"失败"}}`))
}
func successResp(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(200)
	rw.Write([]byte(`{"success":true}`))
}
