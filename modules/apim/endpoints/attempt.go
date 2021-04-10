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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/apim/services/apierrors"
	"github.com/erda-project/erda/modules/apim/services/signauth"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
)

// API 返回对应的错误类型
const (
	ApiTest            = "API_TEST"
	PipelineYmlVersion = "1.1"
	ApiTestType        = "api-test"
	ApiTestIDs         = "api_ids"
	UsecaseID          = "usecase_id"
	PipelineStageLen   = 10
	Project            = "project"
	Usecase            = "case"
)

func (e *Endpoints) ExecuteAttemptTest(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	_, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.UpdateClient.NotLogin().ToResp(), nil
	}

	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.UpdateClient.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	var uriParams = apistructs.AttempTestURIParams{
		AssetID:        vars[urlPathAssetID],
		SwaggerVersion: vars[urlPathSwaggerVersion],
	}

	var body apistructs.APITestReq
	bodyData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return apierrors.AttemptExecuteAPITest.InvalidParameter("invalid body").ToResp(), nil
	}
	if err := json.Unmarshal(bodyData, &body); err != nil {
		return apierrors.AttemptExecuteAPITest.InvalidParameter(err).ToResp(), nil
	}
	if len(body.APIs) == 0 {
		return apierrors.AttemptExecuteAPITest.InvalidParameter(fmt.Errorf("API 个数为 0")).ToResp(), nil
	}
	apiInfo := body.APIs[0]

	// 利用 clientID 查找到 相应的 contract
	var (
		clientModel apistructs.ClientModel
		contract    apistructs.ContractModel
	)
	if err := e.assetSvc.FirstRecord(&clientModel, map[string]interface{}{
		"org_id":    orgID,
		"client_id": body.ClientID,
	}); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return apierrors.AttemptExecuteAPITest.InvalidParameter("无效的 clientID").ToResp(), nil
		}
		return apierrors.AttemptExecuteAPITest.InternalError(err).ToResp(), nil
	}
	if err := e.assetSvc.FirstRecord(&contract, map[string]interface{}{
		"org_id":          orgID,
		"asset_id":        uriParams.AssetID,
		"client_id":       clientModel.ID,
		"swagger_version": uriParams.SwaggerVersion,
		"status":          "proved",
	}); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return apierrors.AttemptExecuteAPITest.InternalError(errors.Wrap(err, "客户端未授权")).ToResp(), nil
		}
		return apierrors.AttemptExecuteAPITest.InternalError(err).ToResp(), nil
	}

	// 查找访问管理
	var access apistructs.APIAccessesModel
	if err = e.assetSvc.FirstRecord(&access, map[string]interface{}{
		"org_id":          orgID,
		"asset_id":        uriParams.AssetID,
		"swagger_version": uriParams.SwaggerVersion,
	}); err != nil {
		return apierrors.AttemptExecuteAPITest.InternalError(errors.Wrap(err, "no access")).ToResp(), nil
	}

	// 查找域名
	domains := e.assetSvc.GetEndpointDomains(access.EndpointID)
	if len(domains) == 0 {
		return apierrors.AttemptExecuteAPITest.InternalError(errors.New("no domain")).ToResp(), nil
	}
	domain := domains[0]

	var request *http.Request
	switch method := strings.ToUpper(apiInfo.Method); method {
	case http.MethodGet, http.MethodHead:
		if request, err = http.NewRequest(method, domain, nil); err != nil {
			return apierrors.AttemptExecuteAPITest.InternalError(errors.New("转发测试请求失败")).ToResp(), nil
		}
	default:
		var (
			content     string
			contentType string
		)
		switch contentType = apiInfo.Body.Type; contentType {
		case "application/json":
			content = string(apiInfo.Body.Content)
		case "application/x-www-form-urlencoded":
			var (
				contentValues = make(url.Values)
				contents      []apistructs.ProxyContent
			)
			if err := json.Unmarshal(apiInfo.Body.Content, &contents); err != nil {
				logrus.Errorf("failed to Unmarshal contents, err: %v", err)
				return apierrors.AttemptExecuteAPITest.InvalidParameter("invalid contents").ToResp(), nil
			}
			for _, v := range contents {
				contentValues.Add(v.Key, v.Value)
			}
			content = contentValues.Encode()
			logrus.Infof("urlEncodedBody: %+v, contents: %v, content: %v", string(apiInfo.Body.Content), contents, content)
		}
		requestBody := bytes.NewBufferString(content)
		if request, err = http.NewRequest(method, domain, requestBody); err != nil {
			return apierrors.AttemptExecuteAPITest.InternalError(errors.New("转发测试请求失败")).ToResp(), nil
		}
		request.Header.Set("content-type", contentType)
	}

	request.Host = domain
	request.URL.Scheme = apiInfo.Schema
	if request.URL.Scheme == "" {
		request.URL.Scheme = "https"
	}
	request.URL.Host = domain
	request.URL.Path = apiInfo.URL
	request.Proto = "HTTP/1.1"
	request.ProtoMajor = 1
	request.ProtoMinor = 1

	// 传入 query parameters
	var params = request.URL.Query()
	for _, v := range apiInfo.Params {
		params.Add(v.Key, v.Value)
	}
	if access.Authentication == apistructs.AuthenticationSignAuth {
		if body.ClientSecret == "" {
			return apierrors.AttemptExecuteAPITest.InvalidParameter("无效的 secret ").ToResp(), nil
		}
		params.Set("appKey", body.ClientID)
	}
	request.URL.RawQuery = params.Encode()

	// 传入 headers
	for _, v := range apiInfo.Header {
		request.Header.Add(v.Key, v.Value)
	}
	if access.Authentication == apistructs.AuthenticationKeyAuth {
		request.Header.Add("X-App-Key", body.ClientID)
	}

	if access.Authentication == apistructs.AuthenticationSignAuth {
		signer, err := signauth.NewSigner(request, body.ClientID, body.ClientSecret)
		if err != nil {
			logrus.Errorf("failed to NewSigner, err: %v", err)
			return apierrors.AttemptExecuteAPITest.InternalError(errors.Errorf("计算 sign-auth 失败: %v", err)).ToResp(), nil
		}
		sign, err := signer.Sign()
		if err != nil {
			logrus.Errorf("failed to Sign, err: %v", err)
			return apierrors.AttemptExecuteAPITest.InternalError(errors.Errorf("计算 sign-auth 失败: %v", err)).ToResp(), nil
		}
		logrus.Debugf("sign-auth: %s", sign)
	}

	if request.Body != nil {
		if data, err := ioutil.ReadAll(request.Body); err == nil {
			logrus.Infof("requesting body: %s", string(data))
			request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		}
	}
	logrus.Infof("requesting: %+v", *request)

	client := *http.DefaultClient
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	response, err := client.Do(request)
	if err != nil {
		logrus.Errorf("failed Do request, request: %+v, err: %v", *request, err)
		return apierrors.AttemptExecuteAPITest.InternalError(errors.Wrap(err, "请求失败")).ToResp(), nil
	}
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return apierrors.AttemptExecuteAPITest.InternalError(errors.Wrap(err, "读取响应失败")).ToResp(), nil
	}
	defer response.Body.Close()

	apiResp := apistructs.APIResp{
		Status:  response.StatusCode,
		Headers: response.Header,
		Body:    content,
		BodyStr: string(content),
	}
	apiReq := apistructs.ProxyAPIRequestInfo{
		Host:    request.URL.Host,
		URL:     request.URL.Path,
		Method:  apiInfo.Method,
		Headers: request.Header,
		Params:  params,
		Body:    apiInfo.Body,
	}

	return httpserver.OkResp(map[string]interface{}{"request": apiReq, "response": apiResp})
}
