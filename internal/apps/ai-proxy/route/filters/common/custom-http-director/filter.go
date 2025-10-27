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
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_style"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/pkg/strutil"
)

type CustomHTTPDirector struct{}

var (
	_ filter_define.ProxyRequestRewriter = (*CustomHTTPDirector)(nil)
)

func New() *CustomHTTPDirector { return &CustomHTTPDirector{} }

func (f *CustomHTTPDirector) OnProxyRequest(pr *httputil.ProxyRequest) error {
	r := pr.Out
	ctx := r.Context()
	provider := ctxhelper.MustGetServiceProvider(ctx)
	providerNormalMeta := metadata.FromProtobuf(provider.Metadata)
	providerMeta := providerNormalMeta.MustToServiceProviderMeta()
	// merge model & provider api segment
	var modelAPISegment *api_segment.API
	modelPublicAPISegmentValue, ok := ctxhelper.MustGetModel(ctx).Metadata.Public["api"]
	if ok {
		modelAPISegment = &api_segment.API{}
		cputil.MustObjJSONTransfer(modelPublicAPISegmentValue, modelAPISegment)
	}

	// handle api style config - merge provider and model configs
	method := pr.In.Method
	pathMatcher := ctxhelper.MustGetPathMatcher(ctx)

	// Merge configs with priority: provider (lower) -> model (higher)
	apiStyleConfig := api_segment.MergeAPIStyleConfig(method, pathMatcher.Pattern, providerMeta.Public.API, modelAPISegment)
	if apiStyleConfig == nil {
		return fmt.Errorf("no APIStyleConfig found, method: %s, path: %s", method, pathMatcher.Pattern)
	}

	// method
	if err := methodDirector(r, *apiStyleConfig); err != nil {
		return fmt.Errorf("failed to set method director, err: %v", err)
	}

	// schema
	if err := schemeDirector(r, *apiStyleConfig); err != nil {
		return fmt.Errorf("failed to set schema director, err: %v", err)
	}

	// host
	if err := hostDirector(r, *apiStyleConfig); err != nil {
		return fmt.Errorf("failed to set host director, err: %v", err)
	}

	// path
	if err := pathDirector(r, *apiStyleConfig); err != nil {
		return fmt.Errorf("failed to set path director, err: %v", err)
	}

	// queryParams
	if err := queryParamsDirector(r, *apiStyleConfig); err != nil {
		return fmt.Errorf("failed to set query params director, err: %v", err)
	}

	// headers
	if err := headersDirector(r, *apiStyleConfig); err != nil {
		return fmt.Errorf("failed to set headers director, err: %v", err)
	}

	// body
	if err := bodyDirector(r, *apiStyleConfig); err != nil {
		return fmt.Errorf("failed to set body director, err: %v", err)
	}

	return nil
}

func methodDirector(r *http.Request, apiStyleConfig api_style.APIStyleConfig) error {
	// default use request method
	method := r.Method
	if apiStyleConfig.Method != "" {
		method = apiStyleConfig.Method
	}
	method = handleJSONPathTemplate(r.Context(), method)
	r.Method = method
	return nil
}

func schemeDirector(r *http.Request, apiStyleConfig api_style.APIStyleConfig) error {
	// default use https
	scheme := "https"
	if apiStyleConfig.Scheme != "" {
		scheme = apiStyleConfig.Scheme
	}
	scheme = handleJSONPathTemplate(r.Context(), scheme)
	r.URL.Scheme = scheme
	return nil
}

func hostDirector(r *http.Request, apiStyleConfig api_style.APIStyleConfig) error {
	// host must be set, otherwise we don't know where to proxy.
	host := apiStyleConfig.Host
	if host == "" {
		return fmt.Errorf("host is empty in APIStyleConfig")
	}
	host = handleJSONPathTemplate(r.Context(), host)
	r.Host = host
	r.URL.Host = host
	r.Header.Set("Host", host)
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

func pathDirector(r *http.Request, apiStyleConfig api_style.APIStyleConfig) error {
	// default use request path
	pathOp := []string{"set", r.URL.Path}
	if apiStyleConfig.Path != nil {
		_pathOp, err := handlePathToPathOp(apiStyleConfig.Path)
		if err != nil {
			return err
		}
		pathOp = _pathOp
	}
	var newPath string
	op := strings.ToLower(pathOp[0])
	switch op {
	case "set":
		newPath = pathOp[1]
	case "replace":
		oldnews := pathOp[1:]
		newPath = strings.NewReplacer(oldnews...).Replace(r.URL.Path)
	default:
		// if the operation is not recognized, we can just ignore it
	}
	newPath = handleJSONPathTemplate(r.Context(), newPath)
	r.URL.Path = newPath
	r.URL.RawPath = ""
	return nil
}

func queryParamsDirector(r *http.Request, apiStyleConfig api_style.APIStyleConfig) error {
	if len(apiStyleConfig.QueryParams) == 0 {
		return nil
	}
	query := r.URL.Query()
	for key, values := range apiStyleConfig.QueryParams {
		if len(values) == 0 {
			continue
		}
		op := strings.ToLower(values[0])
		switch op {
		case "add":
			for _, value := range values[1:] {
				value = handleJSONPathTemplate(r.Context(), value)
				query.Add(key, value)
			}
		case "set":
			for _, value := range values[1:] {
				value = handleJSONPathTemplate(r.Context(), value)
				query.Set(key, value)
			}
		case "delete", "del", "remove":
			query.Del(key)
		default:
			// if the operation is not recognized, we can just ignore it
		}
	}
	r.URL.RawQuery = query.Encode()
	return nil
}

func headersDirector(r *http.Request, apiStyleConfig api_style.APIStyleConfig) error {
	if len(apiStyleConfig.Headers) == 0 {
		return nil
	}
	for key, values := range apiStyleConfig.Headers {
		if len(values) == 0 {
			continue
		}
		op := strings.ToLower(values[0])
		switch op {
		case "add":
			for _, value := range values[1:] {
				value = handleJSONPathTemplate(r.Context(), value)
				r.Header.Add(key, value)
			}
		case "set":
			for _, value := range values[1:] {
				value = handleJSONPathTemplate(r.Context(), value)
				r.Header.Set(key, value)
			}
		case "delete", "del", "remove":
			r.Header.Del(key)
		default:
			// if the operation is not recognized, we can just ignore it
		}
	}
	return nil
}

func handleJSONPathTemplate(ctx context.Context, s string) string {
	parser := api_style.MustNewJSONPathParser(api_style.DefaultRegexpPattern, api_style.DefaultMultiChoiceSplitter)
	if !parser.NeedDoReplace(s) {
		return s
	}
	provider := ctxhelper.MustGetServiceProvider(ctx)
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

func bodyDirector(r *http.Request, apiStyleConfig api_style.APIStyleConfig) error {
	if apiStyleConfig.Body == nil {
		return nil
	}
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		return nil
	}
	transformer := getBodyTransformerByContentType(contentType)
	if transformer == nil {
		return nil
	}
	return transformer.Transform(r, apiStyleConfig.Body)
}
