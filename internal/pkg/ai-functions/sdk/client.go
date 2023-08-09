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

	"github.com/pkg/errors"
)

func NewClient(api *url.URL, client *http.Client) (*Client, error) {
	switch api.Scheme {
	case "http", "https":
	default:
		api.Scheme = "http"
	}
	if api.Host == "" {
		return nil, errors.New("no host in openai url")
	}
	return &Client{api: api, client: client}, nil
}

type Client struct {
	api    *url.URL
	client *http.Client
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

func (c *Client) CreateCompletion(ctx context.Context, req *CreateCompletionOptions) (*ChatCompletions, error) {
	req.Stream = false
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return nil, errors.Wrap(err, "failed to Encode CreateCompletionOptions")
	}
	request, err := http.NewRequest(http.MethodPost, c.URLV1ChatCompletion(), &buf)
	request.Header.Set("Authorization", "Bearer "+"e78c1fe49d704fda978041cb21770282") // todo: ak
	if err != nil {
		return nil, errors.Wrapf(err, "failed to NewRequest, uri: %s", c.URLV1ChatCompletion())
	}
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
	var chatCompletion ChatCompletions
	if err = json.NewDecoder(response.Body).Decode(&chatCompletion); err != nil {
		return nil, errors.Wrap(err, "failed to Decode response to ChatCompletion")
	}

	return &chatCompletion, nil
}
