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

func NewClient(api *url.URL, client *http.Client) *Client {
	return &Client{api: api, client: client}
}

type Client struct {
	api    *url.URL
	client *http.Client
}

func CreateCompletion[FC string | *FunctionCall](ctx context.Context, c *Client, req *CreateCompletionOptions[FC]) (*ChatCompletions, error) {
	if c.api == nil {
		return nil, errors.New("openapi is not specified")
	}
	if c.client == nil {
		c.client = http.DefaultClient
	}
	api := *c.api
	api.Path = "/v1/chat/completions"
	req.Stream = false
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return nil, errors.Wrap(err, "failed to Encode CreateCompletionOptions")
	}
	request, err := http.NewRequest(http.MethodPost, api.String(), &buf)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to NewRequest, uri: %s", api.String())
	}
	response, err := c.client.Do(request)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to Do http request to %s", api.String())
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
