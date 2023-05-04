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

package swagger_ui

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"text/template"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	name = "erda.app.swagger-ui"
)

var (
	providerType = reflect.TypeOf((*provider)(nil))
	spec         = servicehub.Spec{
		Services:    []string{"erda.app.swagger-ui.Server"},
		Summary:     "ai-proxy server",
		Description: "Provides a swagger-ui page",
		ConfigFunc: func() interface{} {
			return new(config)
		},
		Types: []reflect.Type{providerType},
		Creator: func() servicehub.Provider {
			return new(provider)
		},
	}
)

//go:embed index.html
var index string

func init() {
	servicehub.Register(name, &spec)
}

type Interface interface {
	http.Handler

	HandleSwaggerFile(string) func(w http.ResponseWriter, r *http.Request)
	HandleSwaggerName(string) func(w http.ResponseWriter, r *http.Request)
}

type provider struct {
	L logs.Logger
	C *config
	t *template.Template
}

func (p *provider) Init(_ servicehub.Context) (err error) {
	p.L.Infof("[swagger-ui] configuration:\n%s", strutil.TryGetYamlStr(p.C))
	p.t, err = template.ParseGlob(index)
	if err != nil {
		p.L.Fatalf(`failed to template.ParseGlob, err: %v`, err)
	}
	return err
}

func (p *provider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if filename := r.URL.Query().Get("filename"); filename != "" {
		p.HandleSwaggerFile(filename)(w, r)
		return
	}
	if swaggers := len(p.C.Swaggers); swaggers == 0 {
		p.notFound(w, r.URL.Query().Get("provider"))
		return
	}
	if providerName := r.URL.Query().Get("provider"); providerName != "" {
		p.HandleSwaggerName(providerName)(w, r)
		return
	}
	p.redirectToFirstProvider(w, r)
}

func (p *provider) HandleSwaggerName(swaggerName string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if swaggerName == "" {
			p.badRequest(w, "provider name is needed")
			return
		}
		for _, swagger := range p.C.Swaggers {
			if swagger.Name == swaggerName {
				p.HandleSwaggerFile(swagger.Filename)(w, r)
				return
			}
		}
		p.notFound(w, swaggerName)
	}
}

func (p *provider) HandleSwaggerFile(filename string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		p.ui(w, filename)
	}
}

func (p *provider) badRequest(w http.ResponseWriter, err any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": fmt.Sprintf("%v", err),
	})
}

func (p *provider) notFound(w http.ResponseWriter, providerName string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	var m = map[string]any{
		"error":     "provider not found",
		"provider":  providerName,
		"providers": []string{},
	}
	for _, swagger := range p.C.Swaggers {
		m["providers"] = append(m["providers"].([]string), swagger.Name)
	}
	_ = json.NewEncoder(w).Encode(m)
}

func (p *provider) redirectToFirstProvider(w http.ResponseWriter, r *http.Request) {
	if len(p.C.Swaggers) == 0 {
		p.notFound(w, "")
		return
	}
	query := r.URL.Query()
	query.Set("provider", p.C.Swaggers[0].Name)
	r.URL.RawQuery = query.Encode()
	w.Header().Set("Location", r.URL.RequestURI())
	w.WriteHeader(http.StatusPermanentRedirect)
}

func (p *provider) ui(w http.ResponseWriter, filename string) {
	w.Header().Set("Content-Type", "text/html")
	if err := p.t.Execute(w, struct{ SwaggerFilename string }{SwaggerFilename: filename}); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": err.Error(),
		})
	}
}

type config struct {
	Swaggers []swaggerRef `json:"swaggers" yaml:"swaggers"`
}

type swaggerRef struct {
	Name     string `json:"name" yaml:"name"`
	Filename string `json:"filename" yaml:"filename"`
}
