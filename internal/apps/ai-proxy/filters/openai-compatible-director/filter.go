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

package openai_compatible_director

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_style"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

const (
	Name = "openai-compatible-director"
)

var (
	_ reverseproxy.RequestFilter = (*OpenaiCompatibleDirector)(nil)
)

func init() {
	reverseproxy.RegisterFilterCreator(Name, New)
}

type OpenaiCompatibleDirector struct {
	*reverseproxy.DefaultResponseFilter
}

func New(_ json.RawMessage) (reverseproxy.Filter, error) {
	return &OpenaiCompatibleDirector{DefaultResponseFilter: reverseproxy.NewDefaultResponseFilter()}, nil
}

func (f *OpenaiCompatibleDirector) MultiResponseWriter(ctx context.Context) []io.ReadWriter {
	return []io.ReadWriter{ctxhelper.GetLLMDirectorActualResponseBuffer(ctx)}
}

func (f *OpenaiCompatibleDirector) Enable(ctx context.Context, _ *http.Request) bool {
	provider := ctxhelper.MustGetModelProvider(ctx)
	providerNormalMeta := metadata.FromProtobuf(provider.Metadata)
	providerMeta := providerNormalMeta.MustToModelProviderMeta()
	return providerMeta.Public.API != nil &&
		strings.EqualFold(string(providerMeta.Public.API.APIStyle), string(api_style.APIStyleOpenAICompatible))
}

func (f *OpenaiCompatibleDirector) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	provider := ctxhelper.MustGetModelProvider(ctx)
	providerNormalMeta := metadata.FromProtobuf(provider.Metadata)
	providerMeta := providerNormalMeta.MustToModelProviderMeta()

	// handle api style config
	apiStyleConfig := providerMeta.Public.API.APIStyleConfig
	if apiStyleConfig == nil {
		return reverseproxy.Intercept, fmt.Errorf("APIStyleConfig is nil, please check the model provider metadata")
	}

	// method
	if err := methodDirector(ctx, infor, *apiStyleConfig); err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to set method director, err: %v", err)
	}

	// schema
	if err := schemaDirector(ctx, infor, *apiStyleConfig); err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to set schema director, err: %v", err)
	}

	// host
	if err := hostDirector(ctx, infor, *apiStyleConfig); err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to set host director, err: %v", err)
	}

	// path
	if err := pathDirector(ctx, infor, *apiStyleConfig); err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to set path director, err: %v", err)
	}

	// queryParams
	if err := queryParamsDirector(ctx, infor, *apiStyleConfig); err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to set query params director, err: %v", err)
	}

	// headers
	if err := headersDirector(ctx, infor, *apiStyleConfig); err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to set headers director, err: %v", err)
	}

	return reverseproxy.Continue, nil
}

func methodDirector(ctx context.Context, infor reverseproxy.HttpInfor, apiStyleConfig api_style.APIStyleConfig) error {
	method := apiStyleConfig.Method
	if method == "" {
		method = http.MethodPost
	}
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		req.Method = method
	})
	return nil
}

func schemaDirector(ctx context.Context, infor reverseproxy.HttpInfor, apiStyleConfig api_style.APIStyleConfig) error {
	schema := "https"
	if apiStyleConfig.Scheme != "" {
		schema = apiStyleConfig.Scheme
	}
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		req.URL.Scheme = schema
	})
	return nil
}

func hostDirector(ctx context.Context, infor reverseproxy.HttpInfor, apiStyleConfig api_style.APIStyleConfig) error {
	host := apiStyleConfig.Host
	if host == "" {
		return fmt.Errorf("host is empty in APIStyleConfig")
	}
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		req.Host = host
		req.URL.Host = host
		req.Header.Set("Host", host)
		req.Header.Set("X-Forwarded-Host", host)
	})
	return nil
}

func pathDirector(ctx context.Context, infor reverseproxy.HttpInfor, apiStyleConfig api_style.APIStyleConfig) error {
	path := "/v1/chat/completions"
	if apiStyleConfig.Path != "" {
		path = apiStyleConfig.Path
	}
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		req.URL.Path = path
	})
	return nil
}

func queryParamsDirector(ctx context.Context, infor reverseproxy.HttpInfor, apiStyleConfig api_style.APIStyleConfig) error {
	if len(apiStyleConfig.QueryParams) == 0 {
		return nil
	}
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		query := req.URL.Query()
		for key, values := range apiStyleConfig.QueryParams {
			if len(values) == 0 {
				continue
			}
			op := strings.ToLower(values[0])
			switch op {
			case "add":
				for _, value := range values[1:] {
					query.Add(key, value)
				}
			case "set":
				for _, value := range values[1:] {
					query.Set(key, value)
				}
			case "delete", "del", "remove":
				query.Del(key)
			default:
				// if the operation is not recognized, we can just ignore it
			}
		}
		req.URL.RawQuery = query.Encode()
	})
	return nil
}

func headersDirector(ctx context.Context, infor reverseproxy.HttpInfor, apiStyleConfig api_style.APIStyleConfig) error {
	if len(apiStyleConfig.Headers) == 0 {
		return nil
	}
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		for key, values := range apiStyleConfig.Headers {
			if len(values) == 0 {
				continue
			}
			op := strings.ToLower(values[0])
			switch op {
			case "add":
				for _, value := range values[1:] {
					req.Header.Add(key, value)
				}
			case "set":
				for _, value := range values[1:] {
					req.Header.Set(key, value)
				}
			case "delete", "del", "remove":
				req.Header.Del(key)
			default:
				// if the operation is not recognized, we can just ignore it
			}
		}
	})
	return nil
}
