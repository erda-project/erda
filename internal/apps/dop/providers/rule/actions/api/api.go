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
	TemplateParser jsonnet.Engine
}

type API struct {
	URL    string
	Header Header
	// api request body jsonnet snippet
	Snippet string
	// request body jsonnet top level arguments, key: TLA key, value: TLA value
	TLARaw map[string]interface{}
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.TemplateParser = jsonnet.Engine{
		JsonnetVM: gojsonnet.MakeVM(),
	}
	return nil
}

type Header struct {
	Key   string
	Value string
}

// outgoing API
func (a *provider) Send(api *API) (string, error) {
	// parse content with jsonnet
	b, err := json.Marshal(api.TLARaw)
	if err != nil {
		return "", err
	}
	configs := make([]jsonnet.TLACodeConfig, 0)
	configs = append(configs, jsonnet.TLACodeConfig{
		Key:   "ctx",
		Value: string(b),
	})

	jsonStr, err := a.TemplateParser.EvaluateBySnippet(api.Snippet, configs)
	if err != nil {
		return "", err
	}

	// get request body from json string
	res := make(map[string]interface{})
	if err := json.Unmarshal([]byte(jsonStr), &res); err != nil {
		return "", err
	}

	// call outgoing API
	parsed, err := url.Parse(api.URL)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	resp, err := httpclient.New(httpclient.WithHTTPS(), httpclient.WithDialerKeepAlive(time.Duration(a.Cfg.DailerKeepAliveSeconds)*time.Second)).
		Post(parsed.Host).
		Path(parsed.Path).
		Params(parsed.Query()).
		Header("Content-Type", "application/json;charset=utf-8").
		JSONBody(res).Do().
		Body(&buf)
	if err != nil {
		return "", errors.Errorf("Outgoing api: %v, err:%v", api.URL, err)
	}
	if !resp.IsOK() {
		return "", errors.Errorf("Outgoing api: %v, httpcode:%d", api.URL, resp.StatusCode())
	}
	return buf.String(), nil
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
