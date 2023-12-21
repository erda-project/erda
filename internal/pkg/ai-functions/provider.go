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

package ai_functions

import (
	"net/url"
	"os"
	"reflect"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/apps/aifunction/pb"
	"github.com/erda-project/erda/internal/pkg/ai-functions/handler"
)

var (
	_ Handlers = (*provider)(nil)
)

var (
	name = "erda.app.ai-function"
	spec = servicehub.Spec{
		Services: []string{
			"erda.app.ai-function.Handlers",
			"erda.app.ai-function.AIFunction",
		},
		Summary:     "ai function server",
		Description: "AI function server",
		ConfigFunc: func() interface{} {
			return new(Config)
		},
		Types: []reflect.Type{
			reflect.TypeOf((Handlers)(nil)),
			pb.AiFunctionServerType(),
		},
		Creator: func() servicehub.Provider { return new(provider) },
	}
)

func init() {
	servicehub.Register(name, &spec)
}

type provider struct {
	Config *Config
	L      logs.Logger
	R      transport.Register `autowired:"service-register"`
}

func (p *provider) Init(ctx servicehub.Context) (err error) {
	p.Config.openapiURL, err = url.Parse(p.Config.OpenaiAddr)
	if err != nil {
		return err
	}
	_, ok := p.Config.ModelIds["gpt-4"]
	if !ok {
		p.Config.ModelIds["gpt-4"] = os.Getenv(handler.EnvAiProxyChatgpt4ModelId)
	}
	if p.Config.ModelIds["gpt-4"] == "" {
		p.L.Warnf("env %s not set", handler.EnvAiProxyChatgpt4ModelId)
	}

	pb.RegisterAiFunctionImp(p.R, p.AIFunction())
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch ctx.Service() {
	case "erda.app.ai-function.Handlers":
		return Handlers(p)
	case "erda.app.ai-function.AIFunction":
		return p.AIFunction()
	}
	switch ctx.Type() {
	case pb.AiFunctionServerType():
		return p.AIFunction()
	case reflect.TypeOf(Handlers(nil)):
		return Handlers(p)
	}
	return p
}

func (p *provider) AIFunction() pb.AiFunctionServer {
	return &handler.AIFunction{Log: p.L.Sub("AIFunction"), OpenaiURL: p.Config.openapiURL, ModelIds: p.Config.ModelIds}
}

type Config struct {
	// OpenaiAddr is the API address which implemented the OpenAI API.
	// It like https://api.openai.com, https://ai-proxy.erda.cloud
	OpenaiAddr string `json:"openaiAddr" yaml:"openaiAddr"`

	// ModelIds for designated special purposes
	// such as, AI create testcases use chatgpt4 for grouping. ModelIds["chatgpt4"]="5426f79c-****-****-****-1598dc10f1be"
	ModelIds map[string]string `json:"modelIds" yaml:"modelIds"`

	openapiURL *url.URL
}
