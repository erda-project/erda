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

package aiutil

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"

	aiproxyclient "github.com/erda-project/erda/internal/apps/ai-proxy/sdk/client"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/internal/tools/gittar/conf"
	"github.com/erda-project/erda/internal/tools/gittar/models"
)

func InvokeAI(req openai.ChatCompletionRequest, user *models.User, aiSessionID string) string {
	client := getOpenAIClient(user, aiSessionID)
	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		logrus.Warnf("failed to invoke openai, err: %s", err)
		return ""
	}
	if len(resp.Choices) == 0 {
		logrus.Warnf("failed to invoke openai, empty response")
		return ""
	}
	choice := resp.Choices[0]
	if choice.Message.FunctionCall != nil {
		return choice.Message.FunctionCall.Arguments
	}
	return choice.Message.Content
}

func getOpenAIClient(user *models.User, aiSessionID string) *openai.Client {
	// config
	clientConfig := openai.DefaultConfig(aiproxyclient.Instance.Config().ClientAK)
	clientConfig.BaseURL = strings.TrimSuffix(aiproxyclient.Instance.Config().URL, "/") + "/v1"
	clientConfig.HTTPClient = http.DefaultClient
	clientConfig.HTTPClient.Transport = &transport{RoundTripper: http.DefaultTransport, User: user, AISessionID: aiSessionID, ClusterName: conf.DiceCluster()}
	client := openai.NewClientWithConfig(clientConfig)
	return client
}

type transport struct {
	User         *models.User
	AISessionID  string
	ClusterName  string
	RoundTripper http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	m := map[string]string{
		vars.XAIProxySource: "mr-cr" + "___" + t.ClusterName,
	}
	if t.User != nil {
		m[vars.XAIProxyUserId] = t.User.Id
		m[vars.XAIProxyUsername] = t.User.NickName
		m[vars.XAIProxyEmail] = t.User.Email
	}
	for k, v := range m {
		req.Header.Add(k, base64.StdEncoding.EncodeToString([]byte(v)))
	}
	// set api-version
	setAPIVersion := func(version string) {
		key := "api-version"
		query := req.URL.Query()
		query.Del(key)
		query.Set(key, version)
		req.URL.RawQuery = query.Encode()
	}
	setAPIVersion("2023-07-01-preview")
	// set session id
	if t.AISessionID != "" {
		req.Header.Add(vars.XAIProxySessionId, t.AISessionID)
	}
	return t.RoundTripper.RoundTrip(req)
}
