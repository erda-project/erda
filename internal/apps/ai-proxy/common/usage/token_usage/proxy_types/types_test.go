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

package proxy_types

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

func TestDetermineProxyType_NoRequest(t *testing.T) {
	ctx := context.Background()

	if got := DetermineProxyType(ctx); got != ProxyTypeUnknown {
		t.Fatalf("unexpected proxy type, want %s got %s", ProxyTypeUnknown, got)
	}
}

func TestDetermineProxyType(t *testing.T) {
	tests := []struct {
		name string
		path string
		want ProxyType
	}{
		{
			name: "default_openai",
			path: "/v1/chat/completions",
			want: ProxyTypeOpenAI,
		},
		{
			name: "proxy_bailian",
			path: "/proxy/bailian/models",
			want: ProxyTypeProxyBailian,
		},
		{
			name: "proxy_bedrock",
			path: "/proxy/bedrock/models",
			want: ProxyTypeProxyBedrock,
		},
		{
			name: "proxy_unknown",
			path: "/proxy/something",
			want: ProxyTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = ctxhelper.InitCtxMapIfNeed(ctx)

			req := httptest.NewRequest("GET", "http://example.com"+tt.path, nil)
			ctxhelper.PutReverseProxyRequestInSnapshot(ctx, req)

			if got := DetermineProxyType(ctx); got != tt.want {
				t.Fatalf("unexpected proxy type, want %s got %s", tt.want, got)
			}
		})
	}
}
