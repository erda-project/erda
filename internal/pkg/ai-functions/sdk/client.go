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

package sdk

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/pkg/http/httputil"
)

const (
	AIProxyClientAK = "AI_PROXY_CLIENT_AK"
)

func NewClient(api *url.URL, client *http.Client, options ...RequestOption) (*Client, error) {
	switch api.Scheme {
	case "http", "https":
	default:
		api.Scheme = "http"
	}
	if api.Host == "" {
		return nil, errors.New("no host in openai url")
	}
	return &Client{api: api, client: client, options: options}, nil
}

type Client struct {
	api     *url.URL
	client  *http.Client
	options []RequestOption
}

func (c *Client) URLV1ChatCompletion() string {
	var u = *c.api
	u.Path = "/v1/chat/completions"
	return u.String()
}

func (c *Client) HttpClient() *http.Client {
	if c.client == nil {
		return http.DefaultClient
	}
	return c.client
}

func (c *Client) CreateCompletion(ctx context.Context, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	req.Stream = false
	logrus.Infof("CreateCompletion with req: %+v\n", *req)

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return nil, errors.Wrap(err, "failed to Encode CreateCompletionOptions")
	}
	aiProxyClientAK := os.Getenv(AIProxyClientAK)
	if aiProxyClientAK == "" {
		err := errors.Errorf("env %s not set", AIProxyClientAK)
		return nil, errors.Wrap(err, "failed to get ai proxy client ak")
	}
	request, err := http.NewRequest(http.MethodPost, c.URLV1ChatCompletion(), &buf)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to NewRequest, uri: %s", c.URLV1ChatCompletion())
	}
	request.Header.Set(httputil.ContentTypeHeader, "application/json")
	request.Header.Set(httputil.AuthorizationHeader, "Bearer "+aiProxyClientAK)
	for _, o := range c.options {
		o(request)
	}

	logrus.Debugf("Post AI-Proxy Request: %+v", request)
	response, err := c.HttpClient().Do(request)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to Do http request to %s", c.URLV1ChatCompletion())
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != 200 {
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read response body")
		}
		return nil, errors.Errorf("response not ok, status: %s, message: %s", response.Status, string(data))
	}
	var chatCompletion openai.ChatCompletionResponse
	if err = json.NewDecoder(response.Body).Decode(&chatCompletion); err != nil {
		return nil, errors.Wrap(err, "failed to Decode response to ChatCompletion")
	}

	return &chatCompletion, nil
}

type PatchOption func(option *openai.ChatCompletionRequest)

func PathOptionWithModel(model string) PatchOption {
	return func(cco *openai.ChatCompletionRequest) {
		cco.Model = model
	}
}

func PathOptionWithTemperature(temperature float32) PatchOption {
	return func(cco *openai.ChatCompletionRequest) {
		cco.Temperature = temperature
	}
}

type RequestOption func(r *http.Request)

func RequestOptionWithResetAPIVersion(version string) RequestOption {
	return func(r *http.Request) {
		key := "api-version"
		query := r.URL.Query()
		query.Del(key)
		query.Set(key, version)
		r.URL.RawQuery = query.Encode()
	}
}
