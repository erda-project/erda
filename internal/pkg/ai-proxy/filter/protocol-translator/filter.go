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
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/filter"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/provider"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	Name = "protocol-translator"
)

var (
	_ filter.Filter = (*ProtocolTranslator)(nil)
)

func init() {
	filter.Register(Name, New)
}

type ProtocolTranslator struct {
	Config *Config
}

func New(config json.RawMessage) (filter.Filter, error) {
	var cfg Config
	if err := yaml.Unmarshal(config, &cfg); err != nil {
		return nil, err
	}
	return &ProtocolTranslator{Config: &cfg}, nil
}

func (f *ProtocolTranslator) OnHttpRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) (filter.Signal, error) {
	l := ctx.Value(filter.LoggerCtxKey{}).(logs.Logger)
	if err := f.ProcessAll(ctx, r); err != nil {
		l.Errorf(`[ProtocolTranslator].ProcessAll, err: %v`, err)
		return filter.Intercept, err
	}
	return filter.Continue, nil
}

func (f *ProtocolTranslator) OnHttpResponse(ctx context.Context, response *http.Response, r *http.Request) (filter.Signal, error) {
	//return -1, errors.New("测试失败的返回")
	return filter.Continue, nil
}

func (f *ProtocolTranslator) Processors() map[string]any {
	return map[string]any{
		"SetAuthorizationIfNotSpecified":      f.SetAuthorizationIfNotSpecified,
		"SetOpenAIOrganizationIfNotSpecified": f.SetOpenAIOrganizationIfNotSpecified,
		"ReplaceAuthorizationWithAPIKey":      f.ReplaceAuthorizationWithAPIKey,
		"SetAPIKeyIfNotSpecified":             f.SetAPIKeyIfNotSpecified,
		"ReplaceURIPath":                      f.ReplaceURIPath,
		"AddQueries":                          f.AddQueries,
	}
}

func (f *ProtocolTranslator) FindProcessor(ctx context.Context, processor string) (any, error) {
	name, args, err := ParseProcessorNameArgs(processor)
	if err != nil {
		return nil, err
	}
	filter.WithValue(ctx, name, args)
	return f.Processors()[name], nil
}

func (f *ProtocolTranslator) ProcessAll(ctx context.Context, r *http.Request) error {
	var l = ctx.Value(filter.LoggerCtxKey{}).(logs.Logger)
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
	l.Debugf(`[ProtocolTranslator].ProcessAll length of processors found: %v`, len(processors))
	for i := 0; i < len(processors); i++ {
		l.Debugf(`[ProtocolTranslator].ProcessAll processor %s is going to do`, names[i])
		if processors[i] == nil {
			panic("processor is nil")
		}
		if p, ok := processors[i].(func(ctx context.Context, header http.Header)); ok {
			p(ctx, r.Header)
		} else if p, ok := processors[i].(func(ctx context.Context, url *url.URL)); ok {
			p(ctx, r.URL)
		} else {
			panic(fmt.Sprintf("process is invalid type: %T", p))
		}
	}
	return nil
}

func (f *ProtocolTranslator) SetAuthorizationIfNotSpecified(ctx context.Context, header http.Header) {
	prov := ctx.Value(filter.ProviderCtxKey{}).(*provider.Provider)
	if appKey := prov.GetAppKey(); appKey != "" && header.Get("Authorization") == "" {
		header.Set("Authorization", "Bearer "+appKey)
	}
}

func (f *ProtocolTranslator) SetOpenAIOrganizationIfNotSpecified(ctx context.Context, header http.Header) {
	prov := ctx.Value(filter.ProvidersCtxKey{}).(*provider.Provider)
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
	prov := ctx.Value(filter.ProviderCtxKey{}).(*provider.Provider)
	if appKey := prov.GetAppKey(); appKey != "" && header.Get("Api-Key") == "" {
		header.Set("Api-Key", appKey)
	}
}

func (f *ProtocolTranslator) ReplaceURIPath(ctx context.Context, u *url.URL) {
	l := ctx.Value(filter.LoggerCtxKey{}).(logs.Logger)
	s := ctx.Value("ReplaceURIPath").(string)
	s, err := strconv.Unquote(s)
	if err != nil {
		l.Errorf(`[ProtocolTranslator].ReplaceURIPath failed to strconv.Unquote(%s)`, ctx.Value(filter.ReplacedPathCtxKey{}).(string))
		return
	}
	p := ctx.Value(filter.ProviderCtxKey{}).(*provider.Provider)
	for {
		expr, start, end, err := strutil.FirstCustomExpression(s, "${", "}", func(s string) bool {
			s = strings.TrimSpace(s)
			return strings.HasPrefix(s, "env.") || strings.HasPrefix(s, "provider.metadata.")
		})
		if err != nil || start == end {
			break
		}
		if strings.HasPrefix(expr, "env.") {
			s = strutil.Replace(s, os.Getenv(strings.TrimPrefix(expr, ".env")), start, end)
		} else {
			s = strutil.Replace(s, p.Metadata[strings.TrimPrefix(expr, "provider.metadata.")], start, end)
		}
	}
	u.Path = s

	l.Debugf(`[ProtocolTranslator].ReplaceURIPath the url.Path after replaced: %s`, u.Path)
}

func (f *ProtocolTranslator) AddQueries(ctx context.Context, u *url.URL) {
	l := ctx.Value(filter.LoggerCtxKey{}).(logs.Logger)
	s := ctx.Value("AddQueries").(string)
	s, err := strconv.Unquote(s)
	values, err := url.ParseQuery(s)
	if err != nil {
		l.Errorf(`[ProtocolTranslator].AddQueries failed to url.ParseQuery(%s)`, ctx.Value(filter.AddQueriesCtxKey{}).(string))
		return
	}
	queries := u.Query()
	for k, vv := range values {
		for _, v := range vv {
			queries.Add(k, v)
		}
	}
	u.RawQuery = queries.Encode()
	l.Debugf(`[ProtocolTranslator].AddQueries the u.RequestURI() after replaced: %s`, u.RequestURI())
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
