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
	"reflect"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/model_provider"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	Name = "openai-director"
)

var (
	_ reverseproxy.RequestFilter = (*OpenaiDirector)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type OpenaiDirector struct {
	Config *Config

	funcs         map[string]func(ctx context.Context) error
	processorArgs map[string]string
}

func New(config json.RawMessage) (reverseproxy.Filter, error) {
	var cfg Config
	if err := yaml.Unmarshal(config, &cfg); err != nil {
		return nil, err
	}
	return &OpenaiDirector{Config: &cfg}, nil
}

// Enable 检查 request 的 provider.name 是否为 openai, 如是 openai 则启用, 否则不启用
func (f *OpenaiDirector) Enable(ctx context.Context, req *http.Request) bool {
	// 从 context 中取出存储上下文的 map, 从 map 中取出 provider,
	// 这个 provider 由 名为 context 的 filter 插入.
	value, ok := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyModelProvider{})
	if !ok || value == nil {
		return false
	}
	prov, ok := value.(*model_provider.ModelProvider)
	if !ok {
		return false
	}
	return strings.EqualFold(prov.Name, "openai")
}

func (f *OpenaiDirector) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	if err := f.ProcessAll(ctx, infor); err != nil {
		return reverseproxy.Intercept, err
	}
	return reverseproxy.Continue, nil
}

func (f *OpenaiDirector) ProcessAll(ctx context.Context, infor reverseproxy.HttpInfor) error {
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

func (f *OpenaiDirector) FindProcessor(ctx context.Context, processor string) (func(context.Context) error, error) {
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

func (f *OpenaiDirector) DoNothing(context.Context) error { return nil }

func (f *OpenaiDirector) TransAuthorization(ctx context.Context) error {
	value, ok := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyModelProvider{})
	if !ok || value == nil {
		return errors.New("provider not set in context map")
	}
	prov := value.(*model_provider.ModelProvider)
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		req.Header.Del("Authorization")
		req.Header.Set("Authorization", vars.ConcatBearer(prov.APIKey))
	})
	return nil
}

func (f *OpenaiDirector) RewriteScheme(ctx context.Context) error {
	value, ok := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyModelProvider{})
	if !ok || value == nil {
		return errors.New("provider not set in context map")
	}
	prov := value.(*model_provider.ModelProvider)
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		meta, err := prov.Metadata.ToModelProviderMeta()
		if err != nil {
			return
		}
		scheme := meta.Public.Scheme
		if scheme == "http" || scheme == "https" {
			req.URL.Scheme = scheme
		}
		if req.URL.Scheme == "" {
			req.URL.Scheme = "https"
		}
	})
	return nil
}

func (f *OpenaiDirector) RewriteHost(ctx context.Context) error {
	value, ok := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyModelProvider{})
	if !ok || value == nil {
		return errors.New("provider not set in context map")
	}
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		meta, err := (value.(*model_provider.ModelProvider).Metadata.ToModelProviderMeta())
		if err != nil {
			return
		}
		req.Host = meta.Public.Host
		req.URL.Host = req.Host
		req.Header.Set("Host", req.Host)
		req.Header.Set("X-Forwarded-Host", req.Host)
	})
	return nil
}

func (f *OpenaiDirector) AllDirectors() map[string]func(ctx context.Context) error {
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
