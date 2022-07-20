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

package actions

import (
	"bytes"
	"encoding/json"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/internal/core/rule/jsonnet"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type APIConfig struct {
	TemplateParser jsonnet.Engine
	URL            string
	Header         Header
	Body           string
	Snippet        string
}

type Header struct {
	Key   string
	Value string
}

// outgoing API
func (a *APIConfig) Send(content map[string]interface{}) (string, error) {
	b, err := json.Marshal(content)
	if err != nil {
		return "", err
	}
	configs := make([]jsonnet.TLACodeConfig, 0)
	configs = append(configs, jsonnet.TLACodeConfig{
		Key:   "ctx",
		Value: string(b),
	})

	jsonStr, err := a.TemplateParser.EvaluateBySnippet(a.Snippet, configs)
	if err != nil {
		return "", err
	}

	res := make(map[string]interface{})
	if err := json.Unmarshal([]byte(jsonStr), &res); err != nil {
		return "", err
	}
	parsed, err := url.Parse(a.URL)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	resp, err := httpclient.New(httpclient.WithHTTPS(), httpclient.WithDialerKeepAlive(30*time.Second)).
		Post(parsed.Host).
		Path(parsed.Path).
		Params(parsed.Query()).
		Header("Content-Type", "application/json;charset=utf-8").
		JSONBody(res).Do().
		Body(&buf)
	if err != nil {
		return "", errors.Errorf("Outgoing api : %v , err:%v", a.URL, err)
	}
	if !resp.IsOK() {
		return "", errors.Errorf("Outgoing api : %v, httpcode:%d", a.URL, resp.StatusCode())
	}
	return buf.String(), nil
}
