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

package azure_director

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/route"
	"github.com/erda-project/erda/pkg/reverseproxy"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	Name = "azure-director"
)

var (
	_ reverseproxy.RequestFilter = (*AzureDirector)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type AzureDirector struct {
	Config *Config

	funcs         map[string]func(ctx context.Context) error
	processorArgs map[string]string
}

func New(config json.RawMessage) (reverseproxy.Filter, error) {
	var cfg Config
	if err := yaml.Unmarshal(config, &cfg); err != nil {
		return nil, err
	}
	return &AzureDirector{Config: &cfg}, nil
}

// Enable 检查 request 的 provider.name 是否为 azure, 如是 azure 则启用, 否则不启用
func (f *AzureDirector) Enable(ctx context.Context, req *http.Request) bool {
	// 从 context 中取出存储上下文的 map, 从 map 中取出 provider,
	// 这个 provider 由名为 context 的 filter 插入.
	value, ok := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyProvider{})
	if !ok || value == nil {
		return false
	}
	prov, ok := value.(*models.AIProxyProviders)
	if !ok {
		return false
	}
	return strings.EqualFold(prov.Name, "azure")
}

func (f *AzureDirector) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	if err := f.ProcessAll(ctx, infor); err != nil {
		return reverseproxy.Intercept, err
	}
	return reverseproxy.Continue, nil
}

func (f *AzureDirector) ProcessAll(ctx context.Context, infor reverseproxy.HttpInfor) error {
	var l = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger).Sub("ProcessAll")
	var (
		names      []string
		processors []func(context.Context) error
	)
	for _, name := range f.Config.Directors {
		processor, err := f.FindProcessor(ctx, name)
		if err != nil {
			return err
		}
		names = append(names, name)
		processors = append(processors, processor)
	}
	l.Debugf(`%v processors found: %v`, len(processors), names)
	for i := 0; i < len(processors); i++ {
		p := processors[i]
		if p == nil {
			panic("processor is nil")
		}
		if err := p(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (f *AzureDirector) FindProcessor(ctx context.Context, processor string) (func(context.Context) error, error) {
	name, args, err := ParseProcessorNameArgs(processor)
	if err != nil {
		return nil, err
	}
	if f.processorArgs == nil {
		f.processorArgs = make(map[string]string)
	}
	f.processorArgs[name] = args
	return f.AllDirectors()[name], nil
}

func (f *AzureDirector) DoNothing(context.Context) error { return nil }

func (f *AzureDirector) SetAuthorizationIfNotSpecified(ctx context.Context) error {
	value, ok := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyProvider{})
	if !ok || value == nil {
		return errors.New("provider not set in context map")
	}
	prov := value.(*models.AIProxyProviders)
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		if appKey := prov.GetAPIKeyWithDecrypt(); appKey != "" && req.Header.Get("Authorization") == "" {
			req.Header.Set("Authorization", vars.ConcatBearer(appKey))
		}
	})
	return nil
}

// TransAuthorization 将
func (f *AzureDirector) TransAuthorization(ctx context.Context) error {
	value, ok := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyProvider{})
	if !ok || value == nil {
		return errors.New("provider not set in context map")
	}
	prov := value.(*models.AIProxyProviders)
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		req.Header.Set("Api-Key", prov.GetAPIKeyWithDecrypt())
		req.Header.Del("Authorization")
	})
	return nil
}

func (f *AzureDirector) SetAPIKeyIfNotSpecified(ctx context.Context) error {
	value, ok := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyProvider{})
	if !ok || value == nil {
		return errors.New("provider not set in context map")
	}
	prov := value.(*models.AIProxyProviders)
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		if appKey := prov.GetAPIKeyWithDecrypt(); appKey != "" && req.Header.Get("Api-Key") == "" {
			req.Header.Set("Api-Key", appKey)
		}
	})
	return nil
}

func (f *AzureDirector) AddQueries(ctx context.Context) error {
	return f.handleQueries(ctx, "AddQueries")
}

func (f *AzureDirector) SetQueries(ctx context.Context) error {
	return f.handleQueries(ctx, "SetQueries")
}

func (f *AzureDirector) DefaultQueries(ctx context.Context) error {
	return f.handleQueries(ctx, "DefaultQueries")
}

func (f *AzureDirector) RewriteScheme(ctx context.Context) error {
	value, ok := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyProvider{})
	if !ok || value == nil {
		return errors.New("provider not set in context map")
	}
	prov := value.(*models.AIProxyProviders)
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		if prov.Scheme == "http" || prov.Scheme == "https" {
			req.URL.Scheme = prov.Scheme
		}
		if req.URL.Scheme == "" {
			req.URL.Scheme = "https"
		}
	})
	return nil
}

func (f *AzureDirector) RewriteHost(ctx context.Context) error {
	value, ok := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyProvider{})
	if !ok || value == nil {
		return errors.New("provider not set in context map")
	}
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		req.Host = value.(*models.AIProxyProviders).Host
		req.URL.Host = req.Host
		req.Header.Set("Host", req.Host)
		req.Header.Set("X-Forwarded-Host", req.Host)
	})
	return nil
}

func (f *AzureDirector) RewritePath(ctx context.Context) error {
	var l = ctx.Value(reverseproxy.LoggerCtxKey{}).(logs.Logger).Sub("RewritePath")

	rewrite, err := strconv.Unquote(f.processorArgs["RewritePath"])
	if err != nil {
		return errors.Errorf("failed to get RewritePath args, err: %v", err)
	}
	if rewrite == "" {
		return nil
	}

	values := ctx.Value(reverseproxy.CtxKeyPathMatcher{}).(*route.PathMatcher).Values
	prov_, ok := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyProvider{})
	if !ok || prov_ == nil {
		return errors.New("provider not set in context map")
	}
	prov := prov_.(*models.AIProxyProviders)
	var metadata = make(map[string]string)
	if err = json.Unmarshal([]byte(prov.Metadata), &metadata); err != nil {
		l.Warnf("failed to json.Unmarshal prov.Metadata to map[string]string, prov.Name: %s, prov.InstanceID: %s, metadata: %s, err: %v",
			prov.Name, prov.InstanceID, prov.Metadata, err)
	}
	for {
		expr, start, end, err := strutil.FirstCustomExpression(rewrite, "${", "}", func(s string) bool {
			s = strings.TrimSpace(s)
			return strings.HasPrefix(s, "env.") || strings.HasPrefix(s, "provider.metadata.") || strings.HasPrefix(s, "path.")
		})
		if err != nil || start == end {
			break
		}
		if strings.HasPrefix(expr, "env.") {
			rewrite = strutil.Replace(rewrite, os.Getenv(strings.TrimPrefix(expr, "env.")), start, end)
		}
		if strings.HasPrefix(expr, "provider.metadata") && len(metadata) > 0 {
			rewrite = strutil.Replace(rewrite, metadata[strings.TrimPrefix(expr, "provider.metadata.")], start, end)
		}
		if strings.HasPrefix(expr, "path.") {
			rewrite = strutil.Replace(rewrite, values[strings.TrimPrefix(expr, "path.")], start, end)
		}
	}

	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		req.URL.Path = rewrite
	})
	return nil
}

func (f *AzureDirector) AllDirectors() map[string]func(ctx context.Context) error {
	if len(f.funcs) > 0 {
		return f.funcs
	}
	f.funcs = make(map[string]func(ctx context.Context) error)
	typeOf := reflect.TypeOf(f)
	valueOf := reflect.ValueOf(f)
	doNothing, _ := typeOf.MethodByName("DoNothing")
	for i := 0; i < typeOf.NumMethod(); i++ {
		if method := typeOf.Method(i); method.Type == doNothing.Type {
			f.funcs[method.Name] = valueOf.Method(i).Interface().(func(ctx context.Context) error)
		}
	}
	return f.funcs
}

func (f *AzureDirector) responseNotImplementTranslator(w http.ResponseWriter, from, to any) {
	w.Header().Set("Server", "AI Service on Erda")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": "not implement translator",
		"protocols": map[string]any{
			"from": from,
			"to":   to,
		},
	})
}

func (f *AzureDirector) handleQueries(ctx context.Context, funcName string) error {
	s, err := strconv.Unquote(f.processorArgs[funcName])
	if err != nil {
		return err
	}
	values, err := url.ParseQuery(s)
	if err != nil {
		return errors.Wrapf(err, "failed to parse query args: %s", s)
	}
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		queries := req.URL.Query()
		for k, vv := range values {
			for _, v := range vv {
				switch funcName {
				case "SetQueries":
					queries.Set(k, v)
				case "AddQueries":
					queries.Add(k, v)
				case "DefaultQueries":
					if !queries.Has(k) {
						queries.Add(k, v)
					}
				}
			}
		}
		req.URL.RawQuery = queries.Encode()
	})
	return nil
}

type Config struct {
	Directors []string `json:"directors" yaml:"directors"`
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
