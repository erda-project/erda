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

package reverseproxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/erda-project/erda-infra/pkg/transport"
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/pkg/transport/interceptor"
	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	templatepb "github.com/erda-project/erda-proto-go/apps/aiproxy/template/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers"
	"github.com/erda-project/erda/internal/pkg/audit"
)

func TestTrySetAuth_NoAKBehavior(t *testing.T) {
	inter := buildTrySetAuthInterceptor(t)

	t.Run("non-noauth method should reject when no ak", func(t *testing.T) {
		ctx := transport.WithServiceInfo(context.Background(), transport.NewServiceInfo(
			"test",
			audit.GetMethodName(clientpb.ClientServiceServer.Get),
			nil,
		))
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		called := false

		_, err := inter(func(ctx context.Context, req interface{}) (interface{}, error) {
			called = true
			return "ok", nil
		})(ctx, req)
		if err != handlers.ErrAkNotFound {
			t.Fatalf("expected ErrAkNotFound, got %v", err)
		}
		if called {
			t.Fatalf("handler should not be called on auth-required method without ak")
		}
	})

	t.Run("noauth method should pass when no ak", func(t *testing.T) {
		ctx := transport.WithServiceInfo(context.Background(), transport.NewServiceInfo(
			"test",
			audit.GetMethodName(templatepb.TemplateServiceServer.ListModelTemplates),
			nil,
		))
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		called := false

		got, err := inter(func(ctx context.Context, req interface{}) (interface{}, error) {
			called = true
			return "ok", nil
		})(ctx, req)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if !called {
			t.Fatalf("handler should be called for no-auth method without ak")
		}
		if got != "ok" {
			t.Fatalf("unexpected response: %v", got)
		}
	})
}

func buildTrySetAuthInterceptor(t *testing.T) interceptor.Interceptor {
	t.Helper()
	svcOpts := transport.DefaultServiceOptions()
	TrySetAuth(nil)(svcOpts)
	if len(svcOpts.HTTP) == 0 {
		t.Fatalf("expected trySetAuth to inject http options")
	}

	hOpts := &transhttp.HandleOptions{}
	for _, opt := range svcOpts.HTTP {
		opt(hOpts)
	}
	if hOpts.Interceptor == nil {
		t.Fatalf("expected interceptor to be set")
	}
	return hOpts.Interceptor
}
