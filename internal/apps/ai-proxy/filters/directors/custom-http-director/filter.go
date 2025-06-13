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

package custom_http_director

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_style"
	"github.com/erda-project/erda/pkg/reverseproxy"
	"github.com/erda-project/erda/pkg/strutil"
)

type CustomHTTPDirector struct {
	*reverseproxy.DefaultResponseFilter
}

func New() *CustomHTTPDirector {
	return &CustomHTTPDirector{
		DefaultResponseFilter: reverseproxy.NewDefaultResponseFilter(),
	}
}

func (f *CustomHTTPDirector) MultiResponseWriter(ctx context.Context) []io.ReadWriter {
	return []io.ReadWriter{ctxhelper.GetLLMDirectorActualResponseBuffer(ctx)}
}

func (f *CustomHTTPDirector) OnRequest(ctx context.Context, w http.ResponseWriter, infor reverseproxy.HttpInfor) (signal reverseproxy.Signal, err error) {
	provider := ctxhelper.MustGetModelProvider(ctx)
	providerNormalMeta := metadata.FromProtobuf(provider.Metadata)
	providerMeta := providerNormalMeta.MustToModelProviderMeta()

	// handle api style config
	method := infor.Method()
	pathMatcher := ctx.Value(reverseproxy.CtxKeyPath{}).(string)
	apiStyleConfig := providerMeta.Public.API.GetAPIStyleConfigByMethodAndPathMatcher(method, pathMatcher)
	if apiStyleConfig == nil {
		return reverseproxy.Intercept, fmt.Errorf("no APIStyleConfig found, method: %s, path: %s", method, pathMatcher)
	}

	// method
	if err := methodDirector(ctx, infor, *apiStyleConfig); err != nil {
		return reverseproxy.Intercept, fmt.Errorf("failed to set method director, err: %v", err)
	}

	// schema
	if err := schemeDirector(ctx, infor, *apiStyleConfig); err != nil {
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
	// default use request method
	method := infor.Method()
	if apiStyleConfig.Method != "" {
		method = apiStyleConfig.Method
	}
	method = handleJSONPathTemplate(ctx, method)
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		req.Method = method
	})
	return nil
}

func schemeDirector(ctx context.Context, infor reverseproxy.HttpInfor, apiStyleConfig api_style.APIStyleConfig) error {
	// default use https
	scheme := "https"
	if apiStyleConfig.Scheme != "" {
		scheme = apiStyleConfig.Scheme
	}
	scheme = handleJSONPathTemplate(ctx, scheme)
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		req.URL.Scheme = scheme
	})
	return nil
}

func hostDirector(ctx context.Context, infor reverseproxy.HttpInfor, apiStyleConfig api_style.APIStyleConfig) error {
	// host must be set, otherwise we don't know where to proxy.
	host := apiStyleConfig.Host
	if host == "" {
		return fmt.Errorf("host is empty in APIStyleConfig")
	}
	host = handleJSONPathTemplate(ctx, host)
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		req.Host = host
		req.URL.Host = host
		req.Header.Set("Host", host)
		req.Header.Set("X-Forwarded-Host", host)
	})
	return nil
}

func handlePathToPathOp(inputPath any) ([]string, error) {
	switch v := inputPath.(type) {
	case string:
		return []string{"set", v}, nil
	case []byte:
		return []string{"set", string(v)}, nil
	case []string:
		return v, nil
	case []any:
		pathOp := make([]string, 0, len(v))
		for _, item := range v {
			pathOp = append(pathOp, strutil.String(item))
		}
		return pathOp, nil
	default:
		return nil, fmt.Errorf("invalid path: %v", inputPath)
	}
}

func pathDirector(ctx context.Context, infor reverseproxy.HttpInfor, apiStyleConfig api_style.APIStyleConfig) error {
	// default use request path
	pathOp := []string{"set", infor.URL().Path}
	if apiStyleConfig.Path != nil {
		_pathOp, err := handlePathToPathOp(apiStyleConfig.Path)
		if err != nil {
			return err
		}
		pathOp = _pathOp
	}
	reverseproxy.AppendDirectors(ctx, func(req *http.Request) {
		var newPath string
		op := strings.ToLower(pathOp[0])
		switch op {
		case "set":
			newPath = pathOp[1]
		case "replace":
			oldnews := pathOp[1:]
			newPath = strings.NewReplacer(oldnews...).Replace(req.URL.Path)
		default:
			// if the operation is not recognized, we can just ignore it
		}
		newPath = handleJSONPathTemplate(ctx, newPath)
		req.URL.Path = newPath
		req.URL.RawPath = ""
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
					value = handleJSONPathTemplate(ctx, value)
					query.Add(key, value)
				}
			case "set":
				for _, value := range values[1:] {
					value = handleJSONPathTemplate(ctx, value)
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
					value = handleJSONPathTemplate(ctx, value)
					req.Header.Add(key, value)
				}
			case "set":
				for _, value := range values[1:] {
					value = handleJSONPathTemplate(ctx, value)
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

func handleJSONPathTemplate(ctx context.Context, s string) string {
	parser := api_style.MustNewJSONPathParser(api_style.DefaultRegexpPattern, api_style.DefaultMultiChoiceSplitter)
	if !parser.NeedDoReplace(s) {
		return s
	}
	provider := ctxhelper.MustGetModelProvider(ctx)
	model := ctxhelper.MustGetModel(ctx)
	availableObjects := map[string]any{
		"provider": provider,
		"model":    model,
	}
	var jsonMap map[string]any
	cputil.MustObjJSONTransfer(&availableObjects, &jsonMap)
	result := parser.SearchAndReplace(s, jsonMap)
	return result
}
