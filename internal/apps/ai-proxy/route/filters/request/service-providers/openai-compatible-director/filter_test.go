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
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	metadatapb "github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	serviceproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_style"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/router_define/path_matcher"
)

func TestOnProxyRequest_BatchesRoutes(t *testing.T) {
	director := &OpenaiCompatibleDirector{
		CustomHTTPDirector: Creator("", nil).(*OpenaiCompatibleDirector).CustomHTTPDirector,
	}

	testCases := []struct {
		name        string
		method      string
		inURL       string
		pathPattern string
		wantPath    string
	}{
		{
			name:        "create batch",
			method:      http.MethodPost,
			inURL:       "http://origin.local/v1/batches",
			pathPattern: "/v1/batches",
			wantPath:    "/provider/batches/create",
		},
		{
			name:        "list batches",
			method:      http.MethodGet,
			inURL:       "http://origin.local/v1/batches",
			pathPattern: "/v1/batches",
			wantPath:    "/provider/batches/list",
		},
		{
			name:        "retrieve batch",
			method:      http.MethodGet,
			inURL:       "http://origin.local/v1/batches/batch_123",
			pathPattern: "/v1/batches/{batch_id}",
			wantPath:    "/provider/batches/batch_123",
		},
		{
			name:        "cancel batch",
			method:      http.MethodPost,
			inURL:       "http://origin.local/v1/batches/batch_123/cancel",
			pathPattern: "/v1/batches/{batch_id}/cancel",
			wantPath:    "/provider/batches/batch_123/abort",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pr := newProxyRequestWithOpenAIAPIStyle(t, tc.method, tc.inURL, tc.pathPattern)
			err := director.OnProxyRequest(pr)
			require.NoError(t, err)

			require.Equal(t, "https", pr.Out.URL.Scheme)
			require.Equal(t, "api.vendor.test", pr.Out.URL.Host)
			require.Equal(t, "api.vendor.test", pr.Out.Host)
			require.Equal(t, "api.vendor.test", pr.Out.Header.Get("Host"))
			require.Equal(t, tc.wantPath, pr.Out.URL.Path)
		})
	}
}

func TestOnProxyRequest_NoOpWhenNotOpenAICompatible(t *testing.T) {
	director := &OpenaiCompatibleDirector{
		CustomHTTPDirector: Creator("", nil).(*OpenaiCompatibleDirector).CustomHTTPDirector,
	}

	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutPathMatcher(ctx, path_matcher.NewPathMatcher("/v1/batches"))
	ctxhelper.PutServiceProvider(ctx, newServiceProvider(t, string(api_style.APIStyleAnthropicCompatible)))
	ctxhelper.PutModel(ctx, newModel())

	req := httptest.NewRequest(http.MethodPost, "http://origin.local/v1/batches", nil).WithContext(ctx)
	pr := &httputil.ProxyRequest{
		In:  req,
		Out: req.Clone(ctx),
	}
	oldHost := pr.Out.URL.Host
	oldPath := pr.Out.URL.Path

	err := director.OnProxyRequest(pr)
	require.NoError(t, err)
	require.Equal(t, oldHost, pr.Out.URL.Host)
	require.Equal(t, oldPath, pr.Out.URL.Path)
}

func newProxyRequestWithOpenAIAPIStyle(t *testing.T, method, rawURL, pattern string) *httputil.ProxyRequest {
	t.Helper()

	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutPathMatcher(ctx, path_matcher.NewPathMatcher(pattern))
	ctxhelper.PutServiceProvider(ctx, newServiceProvider(t, string(api_style.APIStyleOpenAICompatible)))
	ctxhelper.PutModel(ctx, newModel())

	req := httptest.NewRequest(method, rawURL, nil).WithContext(ctx)
	return &httputil.ProxyRequest{
		In:  req,
		Out: req.Clone(ctx),
	}
}

func newServiceProvider(t *testing.T, style string) *serviceproviderpb.ServiceProvider {
	t.Helper()

	apiValue, err := structpb.NewValue(map[string]any{
		"apiStyle": style,
		"apiStyleConfigs": map[string]any{
			"POST:/v1/batches": map[string]any{
				"host": "api.vendor.test",
				"path": "/provider/batches/create",
			},
			"GET:/v1/batches": map[string]any{
				"host": "api.vendor.test",
				"path": "/provider/batches/list",
			},
			"GET:/v1/batches/{batch_id}": map[string]any{
				"host": "api.vendor.test",
				"path": []any{"replace", "/v1/batches/", "/provider/batches/"},
			},
			"POST:/v1/batches/{batch_id}/cancel": map[string]any{
				"host": "api.vendor.test",
				"path": []any{"replace", "/v1/batches/", "/provider/batches/", "/cancel", "/abort"},
			},
		},
	})
	require.NoError(t, err)

	return &serviceproviderpb.ServiceProvider{
		Metadata: &metadatapb.Metadata{
			Public: map[string]*structpb.Value{
				"api": apiValue,
			},
		},
	}
}

func newModel() *modelpb.Model {
	return &modelpb.Model{
		Metadata: &metadatapb.Metadata{
			Public: map[string]*structpb.Value{},
		},
	}
}
