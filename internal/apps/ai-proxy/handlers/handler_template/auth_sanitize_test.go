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

package handler_template

import (
	"context"
	"testing"

	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	templatepb "github.com/erda-project/erda-proto-go/apps/aiproxy/template/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

func TestSanitizeTemplateListRequestByAuth(t *testing.T) {
	t.Run("anonymous request should be restricted", func(t *testing.T) {
		ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
		req := &templatepb.TemplateListRequest{
			ClientId:       "7d6f4e70-08ba-4ff4-8d6e-1f4f2e3885ad",
			CheckInstance:  true,
			ShowDeprecated: true,
		}

		sanitizeTemplateListRequestByAuth(ctx, req)

		if req.ClientId != "" {
			t.Fatalf("expected ClientId to be cleared for anonymous request, got %q", req.ClientId)
		}
		if req.CheckInstance {
			t.Fatalf("expected CheckInstance to be false for anonymous request")
		}
		if req.ShowDeprecated {
			t.Fatalf("expected ShowDeprecated to be false for anonymous request")
		}
	})

	t.Run("logged-in non-admin should not see deprecated", func(t *testing.T) {
		ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
		ctxhelper.PutClient(ctx, &clientpb.Client{Id: "client-1"})
		req := &templatepb.TemplateListRequest{
			ClientId:       "client-1",
			CheckInstance:  true,
			ShowDeprecated: true,
		}

		sanitizeTemplateListRequestByAuth(ctx, req)

		if !req.CheckInstance {
			t.Fatalf("expected CheckInstance to remain true for logged-in request")
		}
		if req.ClientId != "client-1" {
			t.Fatalf("expected ClientId to remain unchanged, got %q", req.ClientId)
		}
		if req.ShowDeprecated {
			t.Fatalf("expected ShowDeprecated to be false for non-admin request")
		}
	})

	t.Run("admin can keep deprecated and check instance", func(t *testing.T) {
		ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
		ctxhelper.PutIsAdmin(ctx, true)
		req := &templatepb.TemplateListRequest{
			ClientId:       "8c97f830-9ee4-4ac9-96f1-9d7fffd4fa49",
			CheckInstance:  true,
			ShowDeprecated: true,
		}

		sanitizeTemplateListRequestByAuth(ctx, req)

		if req.ClientId == "" {
			t.Fatalf("expected ClientId to remain unchanged for admin")
		}
		if !req.CheckInstance {
			t.Fatalf("expected CheckInstance to remain true for admin")
		}
		if !req.ShowDeprecated {
			t.Fatalf("expected ShowDeprecated to remain true for admin")
		}
	})
}
