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

package reverse_proxy

import (
	"context"
	"encoding/json"
	"net/http"

	"sigs.k8s.io/yaml"

	"github.com/erda-project/erda/internal/pkg/ai-proxy/filter"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/route"
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

func (fil *ProtocolTranslator) OnHttpRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) (filter.Signal, error) {
	if fil.Config.Protocol == "" {
		return filter.Continue, nil
	}
	rout, ok := ctx.Value(filter.RouteCtxKey{}).(*route.Route)
	if !ok {
		w.Header().Set("server", "ai-proxy/erda")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "failed to retrieve route info",
		})
		return filter.Intercept, nil
	}
	if strutil.EqualFold(rout.Protocol, fil.Config.Protocol) {
		return filter.Continue, nil
	}
	fil.responseNotImplementTranslator(w, rout.Protocol, fil.Config.Protocol)
	return filter.Intercept, nil
}

func (fil *ProtocolTranslator) OnHttpResponse(ctx context.Context, response *http.Response, r *http.Request) (filter.Signal, error) {
	//return -1, errors.New("测试失败的返回")
	return filter.Continue, nil
}

func (fil *ProtocolTranslator) responseNotImplementTranslator(w http.ResponseWriter, from, to any) {
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
	Protocol string `json:"protocol" yaml:"protocol"`
}
