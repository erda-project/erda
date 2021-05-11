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
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/mock"
	"github.com/erda-project/erda/pkg/strutil"
)

// APITest API 测试的结构信息
type APITest struct {
	// dice api test 声明，包括入参、出参、断言等
	API *apistructs.APIInfo

	// 运行时结果，包括 response、断言结果、接口测试结果等
	APIResult *apistructs.ApiTestInfo

	opt option
}

func New(api *apistructs.APIInfo, opOptions ...OpOption) *APITest {
	var at APITest
	at.API = api

	for _, opt := range opOptions {
		opt(&at.opt)
	}

	return &at
}

var (
	regexForRender = regexp.MustCompile("{{([^{}]+)}}")

	regexSubmatchReplaceGenerateFun = func(caseParams map[string]*apistructs.CaseParams) func(sub []string) string {
		return func(sub []string) string {
			origin := sub[0]
			inner := sub[1]

			// mock
			if strings.HasPrefix(inner, "@") {
				mockType := strings.TrimPrefix(inner, "@")
				mockValue := mock.MockValue(mockType)
				if mockValue != nil {
					return fmt.Sprint(mockValue)
				}
			}

			// context ref
			if v, ok := caseParams[inner]; ok {
				return fmt.Sprint(v.Value)
			}

			// not found, return origin
			return origin
		}
	}

	renderFunc = func(strWaitRender string, caseParams map[string]*apistructs.CaseParams) string {
		// Use url.PathUnescape other than url.QueryUnescape,
		// avoid to parse time.Time like "2020-12-04T00:00:00+08:00" to "2020-12-04T00:00:00 08:00".
		// PathUnescape is identical to QueryUnescape except that it does not unescape '+' to ' ' (space).
		escapedStr, err := url.PathUnescape(strWaitRender)
		// 若 unescape 成功，则使用转义后的值；否则使用原值
		if err == nil {
			strWaitRender = escapedStr
		}
		return strutil.ReplaceAllStringSubmatchFunc(regexForRender, strWaitRender, regexSubmatchReplaceGenerateFun(caseParams))

	}
)

// Invoke 执行 API 测试
func (at *APITest) Invoke(httpClient *http.Client, testEnv *apistructs.APITestEnvData, caseParams map[string]*apistructs.CaseParams) (
	*apistructs.APIRequestInfo, *apistructs.APIResp, error) {

	// render at once
	if err := at.renderAtOnce(at.API, caseParams); err != nil {
		return nil, nil, err
	}

	// generate api request for invoking
	var apiReq apistructs.APIRequestInfo

	// url
	var domain string
	if testEnv != nil {
		domain = testEnv.Domain
	}
	polishedURL, err := polishURL(at.API.URL, domain)
	if err != nil {
		return nil, nil, err
	}
	apiReq.URL = polishedURL

	// method
	apiReq.Method = at.API.Method

	// params: queryParams and form values
	params := make(url.Values)
	for _, p := range at.API.Params {
		// add to params
		if p.Key != "" {
			params.Add(p.Key, p.Value)
		}
	}
	apiReq.Params = params

	// headers
	if apiReq.Headers == nil {
		apiReq.Headers = make(http.Header)
	}
	if testEnv != nil && testEnv.Header != nil {
		for k, v := range testEnv.Header {
			apiReq.Headers.Set(strings.TrimSpace(k), strings.TrimSpace(v))
		}
	}
	for _, h := range at.API.Headers {
		apiReq.Headers.Set(h.Key, h.Value)
	}

	// request body
	var reqBody string
	if at.API.Body.Content != nil && fmt.Sprint(at.API.Body.Content) != "" {

		switch at.API.Body.Type {
		case apistructs.APIBodyTypeNone:
			reqBody = ""

		case apistructs.APIBodyTypeApplicationXWWWFormUrlencoded:
			values := url.Values{}
			for _, param := range at.API.Body.Content.([]apistructs.APIParam) {
				values.Add(param.Key, param.Value)
			}
			reqBody = values.Encode()

		case apistructs.APIBodyTypeApplicationJSON:
			var reqBodyStr string
			switch at.API.Body.Content.(type) {
			case string:
				reqBodyStr = at.API.Body.Content.(string)
			case []byte:
				reqBodyStr = string(at.API.Body.Content.([]byte))
			default:
				return nil, nil, fmt.Errorf("invalid request body content while body type is application/json, type: %s", reflect.TypeOf(at.API.Body.Content).Kind())
			}
			// check if json is invalid
			var o interface{}
			if err := json.Unmarshal([]byte(reqBodyStr), &o); err != nil {
				// 提前赋值，apiReq 不返回 nil，用于错误时的详情展示
				apiReq.Body.Type = at.API.Body.Type
				apiReq.Body.Content = reqBodyStr
				return &apiReq, nil, fmt.Errorf("failed to json unmarshal request body, value: %s, err: %v", reqBodyStr, err)
			}
			reqBody = reqBodyStr

		default:
			reqBody = fmt.Sprint(at.API.Body.Content)
		}
	}
	apiReq.Body.Type = at.API.Body.Type
	if apiReq.Body.Type != "" && apiReq.Body.Type != apistructs.APIBodyTypeNone {
		apiReq.Headers.Set("Content-Type", apiReq.Body.Type.String())
	}
	apiReq.Body.Content = reqBody

	// use netportal
	customReq, err := handleCustomNetportalRequest(&apiReq, at.opt.netportalOption)
	if err != nil {
		return nil, nil, err
	}

	// polish headers for compression
	apiReq.Headers = polishHeadersForCompression(apiReq.Headers)

	var buffer bytes.Buffer
	req := httpclient.New(httpclient.WithCompleteRedirect()).
		Method(apiReq.Method, customReq.URL.Scheme+"://"+customReq.URL.Host, httpclient.NoRetry).
		Path(customReq.URL.Path).
		Headers(apiReq.Headers)
	httpResp, err := req.Params(apiReq.Params).
		RawBody(bytes.NewBufferString(apiReq.Body.Content.(string))).
		Do().Body(&buffer)
	if err != nil {
		return nil, nil, err
	}

	// resp
	apiResp := apistructs.APIResp{
		Status:  httpResp.StatusCode(),
		Headers: httpResp.Headers(),
		Body:    buffer.Bytes(),
		BodyStr: buffer.String(),
	}

	return &apiReq, &apiResp, nil
}

func (at *APITest) renderAtOnce(apiReq *apistructs.APIInfo, caseParams map[string]*apistructs.CaseParams) error {
	if apiReq == nil {
		return nil
	}
	// url
	apiReq.URL = renderFunc(strings.TrimSpace(apiReq.URL), caseParams)
	// method
	apiReq.Method = renderFunc(strings.TrimSpace(apiReq.Method), caseParams)
	// params
	for i := range apiReq.Params {
		param := apiReq.Params[i]
		apiReq.Params[i].Key = renderFunc(strings.TrimSpace(param.Key), caseParams)
		apiReq.Params[i].Value = renderFunc(strings.TrimSpace(param.Value), caseParams)
	}
	// request body
	if apiReq.Body.Content != nil {
		switch apiReq.Body.Type {
		case apistructs.APIBodyTypeNone:
			// do nothing
		case apistructs.APIBodyTypeText, apistructs.APIBodyTypeTextPlain:
			// check type
			if err := checkBodyType(apiReq.Body, reflect.String); err != nil {
				return err
			}
			// render
			apiReq.Body.Content = renderFunc(apiReq.Body.Content.(string), caseParams)
		case apistructs.APIBodyTypeApplicationJSON, apistructs.APIBodyTypeApplicationJSON2:
			apiReq.Body.Type = apistructs.APIBodyTypeApplicationJSON
			// check type
			kind := reflect.TypeOf(apiReq.Body.Content).Kind()
			var bodyStr string
			switch kind {
			case reflect.String:
				bodyStr = apiReq.Body.Content.(string)
			default:
				// 若不是 string 类型，则 json 序列化处理
				b, err := json.Marshal(apiReq.Body.Content)
				if err != nil {
					return fmt.Errorf("failed to json marshal request body, type: %s, err: %v", kind.String(), err)
				}
				bodyStr = string(b)
			}
			// don't check json before render, maybe quote problems
			// render
			if at.opt.tryV1RenderJsonBodyFirst {
				bodyStr = tryV1RenderRequestBodyStr(bodyStr, caseParams)
			}
			apiReq.Body.Content = renderFunc(bodyStr, caseParams)
		case apistructs.APIBodyTypeApplicationXWWWFormUrlencoded:
			// check type: []apistructs.APIParam
			b, err := json.Marshal(apiReq.Body.Content)
			if err != nil {
				return err
			}
			var content []apistructs.APIParam
			if err := json.Unmarshal(b, &content); err != nil {
				return err
			}
			var renderedContent []apistructs.APIParam
			for i := range content {
				param := content[i]
				param.Key = renderFunc(strings.TrimSpace(param.Key), caseParams)
				param.Value = renderFunc(strings.TrimSpace(param.Value), caseParams)
				param.Desc = renderFunc(strings.TrimSpace(param.Desc), caseParams)
				renderedContent = append(renderedContent, param)
			}
			apiReq.Body.Content = renderedContent
		}
	}
	// headers
	for i := range apiReq.Headers {
		header := apiReq.Headers[i]
		apiReq.Headers[i].Key = renderFunc(strings.TrimSpace(header.Key), caseParams)
		apiReq.Headers[i].Value = renderFunc(strings.TrimSpace(header.Value), caseParams)
	}
	// out params
	for i := range apiReq.OutParams {
		out := apiReq.OutParams[i]
		apiReq.OutParams[i].Key = renderFunc(strings.TrimSpace(out.Key), caseParams)
		apiReq.OutParams[i].Expression = renderFunc(strings.TrimSpace(out.Expression), caseParams)
		apiReq.OutParams[i].Source = apistructs.APIOutParamSource(renderFunc(strings.TrimSpace(out.Source.String()), caseParams))
	}
	// asserts
	for i := range apiReq.Asserts {
		for j := range apiReq.Asserts[i] {
			assert := apiReq.Asserts[i][j]
			apiReq.Asserts[i][j].Arg = renderFunc(strings.TrimSpace(assert.Arg), caseParams)
			apiReq.Asserts[i][j].Operator = renderFunc(strings.TrimSpace(assert.Operator), caseParams)
			apiReq.Asserts[i][j].Value = renderFunc(strings.TrimSpace(assert.Value), caseParams)
		}
	}

	return nil
}
