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

package blacklist_user_agent

import (
	"context"
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

func TestFilter_RejectsBlacklistedUserAgentForClientToken(t *testing.T) {
	t.Cleanup(func() { SetConfig(Config{}) })
	SetConfig(Config{Blacklist: []string{"openclaw"}})

	filter := newFilterForTest(t)
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
		t.Fatal("expected request to be rejected for openclaw")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "forbidden") {
		t.Fatalf("expected forbidden error, got %v", err)
	}
	if got := sink.Snapshot()["blacklist_user_agent"]; got != "openclaw" {
		t.Fatalf("expected blacklist_user_agent note openclaw, got %#v", got)
	}
}

func TestFilter_AllowsAKClientEvenIfUserAgentMatches(t *testing.T) {
	t.Cleanup(func() { SetConfig(Config{}) })
	SetConfig(Config{Blacklist: []string{"openclaw"}})

	filter := newFilterForTest(t)
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

func TestDetectBlacklistedUserAgent_StopsAtFirstMatchedItem(t *testing.T) {
	t.Cleanup(func() { SetConfig(Config{}) })
	SetConfig(Config{Blacklist: []string{"first", "second"}})

	restore := replaceItemsForTest(map[string]BlacklistItem{
		"first": blacklistItemStub{
			name: "first",
			match: func(context.Context) (bool, string) {
				return true, "first-source"
			},
		},
		"second": blacklistItemStub{
			name: "second",
			match: func(context.Context) (bool, string) {
				t.Fatal("expected second blacklist item to be skipped after first match")
				return false, ""
			},
		},
	})
	t.Cleanup(restore)

	gotName, gotSource := detectBlacklistedUserAgent(context.Background())
	if gotName != "first" || gotSource != "first-source" {
		t.Fatalf("expected first match result, got name=%q source=%q", gotName, gotSource)
	}
}

func TestDetectBlacklistedUserAgent_IgnoresUnknownItems(t *testing.T) {
	t.Cleanup(func() { SetConfig(Config{}) })
	SetConfig(Config{Blacklist: []string{"cursor"}})

	gotName, gotSource := detectBlacklistedUserAgent(context.Background())
	if gotName != "" || gotSource != "" {
		t.Fatalf("expected unknown blacklist item to be ignored, got name=%q source=%q", gotName, gotSource)
	}
}

type blacklistItemStub struct {
	name  string
	match func(context.Context) (bool, string)
}

func (s blacklistItemStub) Name() string {
	return s.name
}

func (s blacklistItemStub) Match(ctx context.Context) (bool, string) {
	return s.match(ctx)
}

func newFilterForTest(t *testing.T) filter_define.ProxyRequestRewriter {
	t.Helper()
	return Creator(Name, nil)
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
