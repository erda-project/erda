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

/*
 * All rights Reserved, Designed By Alibaba Group Inc.
 * Copyright: Copyright(C) 1999-2023
 * Company  : Alibaba Group Inc.

 * @brief broadscope completion client
 * @author  yuanci.ytb
 * @version 1.0.0
 * @date 2023-08-04
 */

package bailian_director

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alibabacloud-go/bailian-20230601/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/google/uuid"
)

const (
	BroadscopeBailianPopEndpoint = "bailian.cn-beijing.aliyuncs.com"
	BroadscopeBailianEndpoint    = "https://bailian.aliyuncs.com"
)

type CompletionClient struct {
	TokenData       *client.CreateTokenResponseBodyData
	AccessKeyId     *string
	AccessKeySecret *string
	AgentKey        *string
	PopEndpoint     *string
}

type ChatQaMessage struct {
	User string `json:"User"`
	Bot  string `json:"Bot"`
}

type CompletionRequest struct {
	RequestId        *string            `json:"RequestId"`
	AppId            *string            `json:"AppId"`
	Prompt           *string            `json:"Prompt"`
	SessionId        *string            `json:"SessionId,omitempty"`
	TopP             float32            `json:"TopP,omitempty"`
	Stream           bool               `json:"Stream,omitempty"`
	HasThoughts      bool               `json:"HasThoughts,omitempty"`
	BizParams        map[string]*string `json:"BizParams,omitempty"`
	DocReferenceType *string            `json:"DocReferenceType,omitempty"`
	History          []*ChatQaMessage   `json:"History,omitempty"`
}

type CompletionResponseDataThought struct {
	Thought           *string `json:"Thought,omitempty"`
	ActionType        *string `json:"ActionType,omitempty"`
	ActionName        *string `json:"ActionName,omitempty"`
	Action            *string `json:"Action,omitempty"`
	ActionInputStream *string `json:"ActionInputStream,omitempty"`
	ActionInput       *string `json:"ActionInput,omitempty"`
	Response          *string `json:"Response,omitempty"`
	Observation       *string `json:"Observation,omitempty"`
}

type CompletionResponseDataDocReference struct {
	IndexId *string `json:"IndexId,omitempty"`
	Title   *string `json:"Title,omitempty"`
	DocId   *string `json:"DocId,omitempty"`
	DocName *string `json:"DocName,omitempty"`
	DocUrl  *string `json:"DocUrl,omitempty"`
	Text    *string `json:"Text,omitempty"`
	BizId   *string `json:"BizId,omitempty"`
}

type CompletionResponseData struct {
	ResponseId    *string                               `json:"ResponseId"`
	SessionId     *string                               `json:"SessionId,omitempty"`
	Text          *string                               `json:"Text,omitempty"`
	Thoughts      []*CompletionResponseDataThought      `json:"Thoughts,omitempty"`
	DocReferences []*CompletionResponseDataDocReference `json:"DocReferences,omitempty"`
}

type CompletionResponse struct {
	Success   bool                    `json:"Success"`
	Code      *string                 `json:"Code,omitempty"`
	Message   *string                 `json:"Message,omitempty"`
	RequestId *string                 `json:"RequestId,omitempty"`
	Data      *CompletionResponseData `json:"Data,omitempty"`
}

func (cc *CompletionClient) GetToken() (_token string, _err error) {
	timestamp := time.Now().Unix()
	//Token有效时间24小时, 本地缓存token, 以免每次请求token
	if cc.TokenData == nil || (*cc.TokenData.ExpiredTime-600) < timestamp {
		result, err := cc.CreateToken()
		if err != nil {
			return "", err
		}

		cc.TokenData = result
	}

	return *cc.TokenData.Token, nil
}

func (cc *CompletionClient) CreateToken() (_result *client.CreateTokenResponseBodyData, _err error) {
	if cc.PopEndpoint == nil {
		endpoint := BroadscopeBailianPopEndpoint
		cc.PopEndpoint = &endpoint
	}

	config := &openapi.Config{AccessKeyId: cc.AccessKeyId,
		AccessKeySecret: cc.AccessKeySecret,
		Endpoint:        cc.PopEndpoint}

	tokenClient, err := client.NewClient(config)
	if err != nil {
		return
	}

	request := &client.CreateTokenRequest{AgentKey: cc.AgentKey}
	result, err := tokenClient.CreateToken(request)
	if err != nil {
		return nil, err
	}

	resultBody := result.Body
	if !(*resultBody.Success) {
		return nil, errors.New(*resultBody.Message)
	}

	return resultBody.Data, nil
}

func (cc *CompletionClient) Complete(request *CompletionRequest) (_response *CompletionResponse, _err error) {
	token, err := cc.GetToken()
	if err != nil {
		return nil, err
	}

	if request.RequestId == nil {
		requestId := strings.ReplaceAll(uuid.New().String(), "-", "")
		request.RequestId = &requestId
	}

	url := fmt.Sprintf("%s/v2/app/completions", BroadscopeBailianEndpoint)
	data, err := json.Marshal(*request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	authorization := fmt.Sprintf("Bearer %s", token)

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", authorization)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	response := &CompletionResponse{}
	err = json.Unmarshal(body, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
