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

package protocol_translator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/provider"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	Name = "protocol-translator"
)

var (
	_ reverseproxy.RequestFilter = (*ProtocolTranslator)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type ProtocolTranslator struct {
	Config *Config

	processorArgs map[string]string
}

func New(config json.RawMessage) (reverseproxy.Filter, error) {
	var cfg Config
	if err := yaml.Unmarshal(config, &cfg); err != nil {
		return nil, err
	}
	return &ProtocolTranslator{Config: &cfg}, nil
}

func (f *ProtocolTranslator) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	if err := f.ProcessAll(ctx, infor); err != nil {
		return reverseproxy.Intercept, err
	}
	return reverseproxy.Continue, nil
}

func (f *ProtocolTranslator) ProcessAll(ctx context.Context, infor reverseproxy.HttpInfor) error {
	var l = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger).Sub("ProcessAll")
	var (
		names      []string
		processors []any
	)
	for _, name := range f.Config.Processes {
		processor, err := f.FindProcessor(ctx, name)
		if err != nil {
			return err
		}
		names = append(names, name)
		processors = append(processors, processor)
	}
	l.Debugf(`%v processors found: %v`, len(processors), names)
	for i := 0; i < len(processors); i++ {
		if processors[i] == nil {
			panic("processor is nil")
		}
		switch p := processors[i].(type) {
		case func(ctx context.Context, header http.Header):
			p(ctx, infor.Header())
		case func(ctx context.Context, url *url.URL):
			p(ctx, infor.URL())
		default:
			panic(fmt.Sprintf("process is invalid type: %T", p))
		}
	}
	return nil
}

func (f *ProtocolTranslator) FindProcessor(ctx context.Context, processor string) (any, error) {
	name, args, err := ParseProcessorNameArgs(processor)
	if err != nil {
		return nil, err
	}
	if f.processorArgs == nil {
		f.processorArgs = make(map[string]string)
	}
	f.processorArgs[name] = args
	return f.Processors()[name], nil
}

func (f *ProtocolTranslator) SetAuthorizationIfNotSpecified(ctx context.Context, header http.Header) {
	prov := ctx.Value(reverseproxy.ProviderCtxKey{}).(*provider.Provider)
	if appKey := prov.GetAppKey(); appKey != "" && header.Get("Authorization") == "" {
		header.Set("Authorization", "Bearer "+appKey)
	}
}

func (f *ProtocolTranslator) SetOpenAIOrganizationIfNotSpecified(ctx context.Context, header http.Header) {
	prov := ctx.Value(reverseproxy.ProvidersCtxKey{}).(*provider.Provider)
	if org := prov.GetOrganization(); org != "" && header.Get("OpenAI-Organization") == "" {
		header.Set("OpenAI-Organization", org)
	}
}

func (f *ProtocolTranslator) ReplaceAuthorizationWithAPIKey(ctx context.Context, header http.Header) {
	v := header.Get("Authorization")
	v = strings.TrimPrefix(v, "Bearer ")
	header.Set("Api-Key", v)
	header.Del("Authorization")
}

func (f *ProtocolTranslator) SetAPIKeyIfNotSpecified(ctx context.Context, header http.Header) {
	prov := ctx.Value(reverseproxy.ProviderCtxKey{}).(*provider.Provider)
	if appKey := prov.GetAppKey(); appKey != "" && header.Get("Api-Key") == "" {
		header.Set("Api-Key", appKey)
	}
}

func (f *ProtocolTranslator) AddQueries(ctx context.Context, u *url.URL) {
	l := ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger).Sub("AddQueries")
	s, err := strconv.Unquote(f.processorArgs["AddQueries"])
	values, err := url.ParseQuery(s)
	if err != nil {
		l.Errorf(`AddQueries failed to url.ParseQuery(%s)`, ctx.Value(reverseproxy.AddQueriesCtxKey{}).(string))
		return
	}
	queries := u.Query()
	for k, vv := range values {
		for _, v := range vv {
			queries.Add(k, v)
		}
	}
	u.RawQuery = queries.Encode()
}

func (f *ProtocolTranslator) Processors() map[string]any {
	return map[string]any{
		"SetAuthorizationIfNotSpecified":      f.SetAuthorizationIfNotSpecified,
		"SetOpenAIOrganizationIfNotSpecified": f.SetOpenAIOrganizationIfNotSpecified,
		"ReplaceAuthorizationWithAPIKey":      f.ReplaceAuthorizationWithAPIKey,
		"SetAPIKeyIfNotSpecified":             f.SetAPIKeyIfNotSpecified,
		"AddQueries":                          f.AddQueries,
	}
}

func (f *ProtocolTranslator) responseNotImplementTranslator(w http.ResponseWriter, from, to any) {
	w.Header().Set("server", "ai-proxy/erda")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": "not implement translator",
		"protocols": map[string]any{
			"from": from,
			"to":   to,
		},
	})
}

type Config struct {
	Processes []string
}

func ParseProcessorNameArgs(s string) (string, string, error) {
	index := strings.IndexByte(s, '(')
	if index < 0 {
		return s, "", nil
	}
	lastIndex := strings.LastIndexByte(s, ')')
	if index < 0 {
		return "", "", errors.Errorf("failed to ParseProcessorNameArgs, the configuration %s may be invalid", s)
	}
	if index+1 > lastIndex {
		return "", "", errors.Errorf("failed to ParseProcessorNameArgs, the configuration %s may be invalid", s)
	}
	name, args := s[:index], s[index+1:lastIndex]
	return name, args, nil
}
