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

package forbidden_client_detector

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	clienttokenpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_token/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/message"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
)

func TestFilter_RejectsOpenClawClientTokenByMessageGroup(t *testing.T) {
	filter := newFilterForTest(t, map[string]any{
		"blacklist": []string{"openclaw"},
	})
	pr, sink := newProxyRequestForTest()
	ctxhelper.PutClientToken(pr.In.Context(), &clienttokenpb.ClientToken{Token: "t_test"})
	ctxhelper.PutMessageGroup(pr.In.Context(), message.Group{
		RequestedMessages: message.Messages{
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a personal assistant running inside OpenClaw",
			},
		},
	})

	err := filter.OnProxyRequest(pr)
	if err == nil {
		t.Fatal("expected request to be rejected for openclaw client token")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "forbidden") {
		t.Fatalf("expected forbidden error, got %v", err)
	}
	if got := sink.Snapshot()["forbidden_client"]; got != "openclaw" {
		t.Fatalf("expected forbidden_client note openclaw, got %#v", got)
	}
}

func TestFilter_AllowsAKClientEvenIfPromptMatches(t *testing.T) {
	filter := newFilterForTest(t, map[string]any{
		"blacklist": []string{"openclaw"},
	})
	pr, _ := newProxyRequestForTest()
	ctxhelper.PutClient(pr.In.Context(), &clientpb.Client{Id: "c1"})
	ctxhelper.PutMessageGroup(pr.In.Context(), message.Group{
		RequestedMessages: message.Messages{
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a personal assistant running inside OpenClaw",
			},
		},
	})

	if err := filter.OnProxyRequest(pr); err != nil {
		t.Fatalf("expected ak client request to pass, got %v", err)
	}
}

func TestFilter_UsesEnvBlacklistOverride(t *testing.T) {
	t.Setenv(envKeyBlacklist, "openclaw")

	filter := newFilterForTest(t, nil)
	pr, _ := newProxyRequestForTest()
	ctxhelper.PutClientToken(pr.In.Context(), &clienttokenpb.ClientToken{Token: "t_test"})
	ctxhelper.MustGetAuditSink(pr.In.Context()).Note("prompt", "You are a personal assistant running inside OpenClaw")

	if err := filter.OnProxyRequest(pr); err == nil {
		t.Fatal("expected request to be rejected when env blacklist enables openclaw")
	}
}

func TestFilter_IgnoresUnsupportedBlacklistBranch(t *testing.T) {
	filter := newFilterForTest(t, map[string]any{
		"blacklist": []string{"cursor"},
	})
	pr, _ := newProxyRequestForTest()
	ctxhelper.PutClientToken(pr.In.Context(), &clienttokenpb.ClientToken{Token: "t_test"})
	ctxhelper.PutMessageGroup(pr.In.Context(), message.Group{
		RequestedMessages: message.Messages{
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a personal assistant running inside OpenClaw",
			},
		},
	})

	if err := filter.OnProxyRequest(pr); err != nil {
		t.Fatalf("expected unsupported blacklist branch to be ignored, got %v", err)
	}
}

func newFilterForTest(t *testing.T, cfg map[string]any) filter_define.ProxyRequestRewriter {
	t.Helper()
	var raw json.RawMessage
	if cfg != nil {
		var err error
		raw, err = json.Marshal(cfg)
		if err != nil {
			t.Fatalf("failed to marshal config: %v", err)
		}
	}
	return Creator(Name, raw)
}

func newProxyRequestForTest() (*httputil.ProxyRequest, types.Sink) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutLogger(ctx, logrusx.New())
	sink := types.New("audit-1", logrusx.New())
	ctxhelper.PutAuditSink(ctx, sink)

	req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/chat/completions", nil).WithContext(ctx)
	outReq := req.Clone(ctx)
	return &httputil.ProxyRequest{In: req, Out: outReq}, sink
}
