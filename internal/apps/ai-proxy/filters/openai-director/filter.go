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

package openai_director

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda-infra/base/logs"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	modelproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/http/httputil"
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
	*reverseproxy.DefaultResponseFilter
	Config *Config

	funcs         map[string]func(ctx context.Context) error
	processorArgs map[string]string
}

func New(config json.RawMessage) (reverseproxy.Filter, error) {
	var cfg Config
	if err := yaml.Unmarshal(config, &cfg); err != nil {
		return nil, err
	}
	return &OpenaiDirector{DefaultResponseFilter: reverseproxy.NewDefaultResponseFilter(), Config: &cfg}, nil
}

func (f *OpenaiDirector) MultiResponseWriter(ctx context.Context) []io.ReadWriter {
	return []io.ReadWriter{ctxhelper.GetLLMDirectorActualResponseBuffer(ctx)}
}

// Enable 检查 request 的 provider.name 是否为 openai, 如是 openai 则启用, 否则不启用
func (f *OpenaiDirector) Enable(ctx context.Context, req *http.Request) bool {
	// 从 context 中取出存储上下文的 map, 从 map 中取出 provider,
	// 这个 provider 由 名为 context 的 filter 插入.
	value, ok := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyModelProvider{})
	if !ok || value == nil {
		return false
	}
	prov, ok := value.(*modelproviderpb.ModelProvider)
	if !ok {
		return false
	}
	return prov.Type == modelproviderpb.ModelProviderType_OpenAI.String()
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
	prov := value.(*modelproviderpb.ModelProvider)
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		req.Header.Del(httputil.HeaderKeyAuthorization)
		// ensure auth type
		authValue := prov.ApiKey
		if !strings.Contains(authValue, " ") {
			// if not contains space, treat as no auth type specified, use Bearer as default for backward compatibility
			authValue = vars.ConcatBearer(authValue)
		}
		req.Header.Set(httputil.HeaderKeyAuthorization, authValue)
	})
	return nil
}

func (f *OpenaiDirector) RewriteScheme(ctx context.Context) error {
	value, ok := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyModelProvider{})
	if !ok || value == nil {
		return errors.New("provider not set in context map")
	}
	prov := value.(*modelproviderpb.ModelProvider)
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		meta := metadata.FromProtobuf(prov.Metadata)
		mpMeta, err := meta.ToModelProviderMeta()
		if err != nil {
			return
		}
		scheme := mpMeta.Public.Scheme
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
		meta := metadata.FromProtobuf(value.(*modelproviderpb.ModelProvider).Metadata)
		mpMeta, err := meta.ToModelProviderMeta()
		if err != nil {
			return
		}
		req.Host = mpMeta.Public.Host
		req.URL.Host = req.Host
		req.Header.Set("Host", req.Host)
		req.Header.Set("X-Forwarded-Host", req.Host)
	})
	return nil
}

func (f *OpenaiDirector) AddModelInRequestBody(ctx context.Context) error {
	value, ok := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyModel{})
	if !ok || value == nil {
		return errors.New("model not set in context map")
	}
	model := value.(*modelpb.Model)
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		infor := reverseproxy.NewInfor(ctx, req)
		// read body to json, then add a `model` field, then write back to body
		var body map[string]interface{}
		if err := json.NewDecoder(infor.BodyBuffer()).Decode(&body); err != nil && err != io.EOF {
			ctxhelper.GetLogger(ctx).Errorf("failed to decode request body, err: %v", err)
			return
		}
		body["model"] = model.Name
		b, err := json.Marshal(body)
		if err != nil {
			ctxhelper.GetLogger(ctx).Errorf("failed to marshal request body, err: %v", err)
			return
		}
		infor.SetBody(io.NopCloser(strings.NewReader(string(b))), int64(len(b)))
	})
	return nil
}

func (f *OpenaiDirector) AddContextMessages(ctx context.Context) error {
	messageGroup, ok := ctxhelper.GetMessageGroup(ctx)
	if !ok {
		return nil
	}
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		infor := reverseproxy.NewInfor(ctx, req)
		var openaiReq openai.ChatCompletionRequest

		// init `JSONSchema.Schema` for `json.Decode`, otherwise, it will report an error
		openaiReq.ResponseFormat = &openai.ChatCompletionResponseFormat{
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Schema: &jsonschema.Definition{},
			},
		}
		if err := json.NewDecoder(infor.BodyBuffer()).Decode(&openaiReq); err != nil && err != io.EOF {
			ctxhelper.GetLogger(ctx).Errorf("failed to decode request body, err: %v", err)
			return
		}
		if openaiReq.ResponseFormat.Type == "" {
			openaiReq.ResponseFormat.Type = openai.ChatCompletionResponseFormatTypeText
		}
		if openaiReq.ResponseFormat.Type != openai.ChatCompletionResponseFormatTypeJSONSchema {
			openaiReq.ResponseFormat.JSONSchema = nil
		}
		openaiReq.Messages = messageGroup.AllMessages
		b, err := json.Marshal(&openaiReq)
		if err != nil {
			ctxhelper.GetLogger(ctx).Errorf("failed to marshal request body, err: %v", err)
			return
		}
		infor.SetBody(io.NopCloser(strings.NewReader(string(b))), int64(len(b)))
	})
	return nil
}

func (f *OpenaiDirector) AddHeaders(ctx context.Context) error {
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		headerKVs := strings.TrimSpace(f.processorArgs["AddHeaders"])
		// split by comma: a=b,c=d
		kvs := strings.Split(headerKVs, ",")
		for _, kv := range kvs {
			// split by =
			parts := strings.Split(kv, "=")
			if len(parts) != 2 {
				continue
			}
			// trim space
			key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			req.Header.Set(key, value)
		}
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
	if lastIndex < 0 {
		return "", "", errors.Errorf("failed to ParseProcessorNameArgs, the configuration %s may be invalid", s)
	}
	if index+1 > lastIndex {
		return "", "", errors.Errorf("failed to ParseProcessorNameArgs, the configuration %s may be invalid", s)
	}
	name, args := s[:index], s[index+1:lastIndex]
	return name, args, nil
}
