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

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
	"time"

	gojsonnet "github.com/google/go-jsonnet"
	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/internal/apps/dop/providers/rule/jsonnet"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type config struct {
	DailerKeepAliveSeconds int64 `default:"30" file:"DIALER_KEEP_ALIVE_SECONDS" env:"DIALER_KEEP_ALIVE_SECONDS"`
}

type Interface interface {
	Send(api *API) (string, error)
}

type provider struct {
	Cfg            *config
	TemplateParser *jsonnet.Engine
}

type API struct {
	URL string
	// api request config jsonnet snippet
	Snippet string
	// request jsonnet top level arguments, key: TLA key, value: TLA value
	TLARaw map[string]interface{}
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.TemplateParser = &jsonnet.Engine{
		JsonnetVM: gojsonnet.MakeVM(),
	}
	return nil
}

type APIConfig struct {
	Method  string                 `json:"method"`
	URL     string                 `json:"url"`
	Headers http.Header            `json:"header"`
	Body    map[string]interface{} `json:"body"`
}

// outgoing API
func (a *provider) Send(api *API) (string, error) {
	apiConfig, err := a.getAPIConfig(api)
	if err != nil {
		return "", err
	}

	parsed, err := url.Parse(apiConfig.URL)
	if err != nil {
		return "", err
	}

	// call outgoing API
	req := httpclient.New(httpclient.WithDialerKeepAlive(time.Duration(a.Cfg.DailerKeepAliveSeconds)*time.Second)).
		Method(apiConfig.Method, parsed.Host).
		Path(parsed.Path).
		Params(parsed.Query()).
		Headers(apiConfig.Headers).
		Header("Content-Type", "application/json;charset=utf-8").
		JSONBody(apiConfig.Body)

	var buf bytes.Buffer
	resp, err := req.Do().Body(&buf)
	responseBody := buf.String()
	if err != nil {
		return "", errors.Errorf("Outgoing api: %v, err:%v, body:%v", api.URL, err, responseBody)
	}
	if !resp.IsOK() {
		return "", errors.Errorf("Outgoing api: %v, httpcode:%d, body:%v", api.URL, resp.StatusCode(), responseBody)
	}
	return responseBody, nil
}

func (a *provider) getAPIConfig(api *API) (*APIConfig, error) {
	// parse content with jsonnet
	b, err := json.Marshal(api.TLARaw)
	if err != nil {
		return nil, err
	}
	configs := make([]jsonnet.TLACodeConfig, 0)
	configs = append(configs, jsonnet.TLACodeConfig{
		Key:   "ctx",
		Value: string(b),
	})

	jsonStr, err := a.TemplateParser.EvaluateBySnippet(api.Snippet, configs)
	if err != nil {
		return nil, err
	}

	// get API config from json string
	var apiConfig APIConfig
	if err := json.Unmarshal([]byte(jsonStr), &apiConfig); err != nil {
		return nil, err
	}
	if apiConfig.URL == "" {
		apiConfig.URL = api.URL
	}
	return &apiConfig, nil
}

func init() {
	servicehub.Register("erda.dop.rule.action.api", &servicehub.Spec{
		Services:   []string{"erda.core.rule.action.api"},
		Types:      []reflect.Type{reflect.TypeOf((*Interface)(nil)).Elem()},
		ConfigFunc: func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
