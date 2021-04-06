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

package apitestsv2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/jsonpath"
)

// ParseOutParams 解析 API 执行结果的出参，存储为全局变量，供后续使用
func (at *APITest) ParseOutParams(apiOutParams []apistructs.APIOutParam, apiResp *apistructs.APIResp,
	caseParams map[string]*apistructs.CaseParams) map[string]interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered from ", r)
		}
	}()
	outParams := make(map[string]interface{})
	jqOrJsonPath := ""
	for _, t := range apiOutParams {
		pam := &apistructs.CaseParams{}
		switch t.Source {
		case apistructs.APIOutParamSourceStatus:
			pam.Type = apistructs.APIOutParamSourceStatus.String()
			pam.Value = apiResp.Status
		case apistructs.APIOutParamSourceBodyJsonJsonPath:
			jqOrJsonPath = "jsonpath"
			fallthrough
		case apistructs.APIOutParamSourceBodyJsonJQ:
			jqOrJsonPath = "jq"
			fallthrough
		case apistructs.APIOutParamSourceBodyJson:
			var (
				body interface{}
				val  interface{}
			)
			// 出参取不到值，为空，忽略错误
			d := json.NewDecoder(bytes.NewReader(apiResp.Body))
			d.UseNumber()
			err := d.Decode(&body)
			if err != nil {
				continue
			}

			express := strings.TrimSpace(t.Expression)
			if express != "" {
				switch jqOrJsonPath {
				case "jsonpath":
					val, _ = jsonpath.Get(body, express)
				case "jq":
					val, _ = jsonpath.JQ(apiResp.BodyStr, express)
				default:
					var err error
					val, err = jsonpath.JQ(apiResp.BodyStr, express)
					if err != nil {
						val, _ = jsonpath.Get(body, express)
					}
				}
			} else {
				val = body
			}

			pam.Type = apistructs.APIOutParamSourceStatus.String()
			pam.Value = val
		case apistructs.APIOutParamSourceBodyText:
			pam.Type = apistructs.APIOutParamSourceStatus.String()
			pam.Value = fmt.Sprint(apiResp.Body)
		case apistructs.APIOutParamSourceHeader:
			pam.Type = apistructs.APIOutParamSourceStatus.String()
			express := strings.TrimSpace(t.Expression)
			if express == "" {
				pam.Value = apiResp.Headers
				break
			}
			if val, ok := apiResp.Headers[express]; ok {
				for i, h := range val {
					if i == 0 {
						pam.Value = h
						continue
					}
					pam.Value = fmt.Sprint(pam.Value, ",", h)
				}
			}
		}

		if t.Key != "" {
			outParams[t.Key] = pam.Value

			// store to case params
			caseParams[t.Key] = pam
		}
	}

	return outParams
}
