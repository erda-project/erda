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
	"testing"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	clienttokenpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_token/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/message"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
)

func TestFilter_RejectsBlacklistedUserAgentForClientTokenFromAuditPrompt(t *testing.T) {
	t.Cleanup(func() { SetConfig(Config{}) })
	SetConfig(Config{
		ClientToken: ClientTokenConfig{Blacklist: []string{"openclaw"}},
	})

	filter := newFilterForTest(t)
	pr, sink := newProxyRequestForTest()
	ctxhelper.PutClientToken(pr.In.Context(), &clienttokenpb.ClientToken{Token: "t_test"})
	audithelper.Note(pr.In.Context(), "prompt", openClawSystemPromptHint+"\n## Tooling\nTool availability")

	err := filter.OnProxyRequest(pr)
	if err == nil {
		t.Fatal("expected request to be rejected for openclaw")
	}
	if got := sink.Snapshot()["blacklist_user_agent"]; got != "openclaw" {
		t.Fatalf("expected blacklist_user_agent note openclaw, got %#v", got)
	}
	if got := sink.Snapshot()["blacklist_user_agent_match_source"]; got != sourceAuditPrompt {
		t.Fatalf("expected blacklist_user_agent_match_source %q, got %#v", sourceAuditPrompt, got)
	}
}

func TestFilter_RejectsAKClientWhenClientBlacklistConfiguredFromMessageGroup(t *testing.T) {
	t.Cleanup(func() { SetConfig(Config{}) })
	SetConfig(Config{
		Client: ClientConfig{Blacklist: []string{"openclaw"}},
	})

	filter := newFilterForTest(t)
	pr, _ := newProxyRequestForTest()
	ctxhelper.PutClient(pr.In.Context(), &clientpb.Client{Id: "c1"})
	ctxhelper.PutMessageGroup(pr.In.Context(), message.Group{
		RequestedMessages: message.Messages{
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: openClawSystemPromptHint + "\n## Tooling\nTool availability",
			},
		},
	})

	if err := filter.OnProxyRequest(pr); err == nil {
		t.Fatal("expected ak client request to be rejected when client blacklist enables openclaw")
	}
}

func TestFilter_RejectsGeneralItemWhenConfiguredFromAuditPrompt(t *testing.T) {
	t.Cleanup(func() { SetConfig(Config{}) })
	SetConfig(Config{
		ClientToken: ClientTokenConfig{Blacklist: []string{"general"}},
		General:     GeneralConfig{ItemTypes: []string{"claude code"}},
	})

	filter := newFilterForTest(t)
	pr, sink := newProxyRequestForTest()
	ctxhelper.PutClientToken(pr.In.Context(), &clienttokenpb.ClientToken{Token: "t_test"})
	audithelper.Note(pr.In.Context(), "prompt", "You are Claude Code, Anthropic's official CLI for Claude.")

	err := filter.OnProxyRequest(pr)
	if err == nil {
		t.Fatal("expected request to be rejected for configured general item")
	}
	if got := sink.Snapshot()["blacklist_user_agent"]; got != "general" {
		t.Fatalf("expected blacklist_user_agent note general, got %#v", got)
	}
}

func TestFilter_AllowsGeneralItemWhenNoConfiguredItemTypes(t *testing.T) {
	t.Cleanup(func() { SetConfig(Config{}) })
	SetConfig(Config{
		ClientToken: ClientTokenConfig{Blacklist: []string{"general"}},
	})

	filter := newFilterForTest(t)
	pr, _ := newProxyRequestForTest()
	ctxhelper.PutClientToken(pr.In.Context(), &clienttokenpb.ClientToken{Token: "t_test"})
	audithelper.Note(pr.In.Context(), "prompt", "You are Claude Code, Anthropic's official CLI for Claude.")

	if err := filter.OnProxyRequest(pr); err != nil {
		t.Fatalf("expected request to pass when general item has no configured item types, got %v", err)
	}
}

func TestResolveEnabledItems_IgnoresUnknownItemsAndPreservesOrder(t *testing.T) {
	restore := replaceItemsForTest(map[string]BlacklistItem{
		"cursor":   namedStubItem{name: "cursor"},
		"openclaw": namedStubItem{name: "openclaw"},
	})
	t.Cleanup(restore)

	items := resolveEnabledItems([]string{"cursor", "unknown", "openclaw"})
	if len(items) != 2 {
		t.Fatalf("expected 2 enabled items, got %d", len(items))
	}
	if items[0].Name() != "cursor" || items[1].Name() != "openclaw" {
		t.Fatalf("expected enabled items to preserve blacklist order, got %q then %q", items[0].Name(), items[1].Name())
	}
}

func TestResolveEnabledItems_SkipsDisabledItems(t *testing.T) {
	restore := replaceItemsForTest(map[string]BlacklistItem{
		"enabled":  namedStubItem{name: "enabled"},
		"disabled": disabledStubItem{name: "disabled"},
	})
	t.Cleanup(restore)

	items := resolveEnabledItems([]string{"enabled", "disabled"})
	if len(items) != 1 || items[0].Name() != "enabled" {
		t.Fatalf("expected disabled items to be skipped, got %#v", items)
	}
}

func TestDetectBlacklistedUserAgent_PrefersHeaderBeforePrompt(t *testing.T) {
	restore := replaceItemsForTest(map[string]BlacklistItem{
		"prompt-first": testPromptItem{name: "prompt-first", match: true},
		"header-last":  testHeaderItem{name: "header-last", match: true},
	})
	t.Cleanup(restore)

	signals := PreparedSignals{
		HeaderPairs: []HeaderPair{{Key: "User-Agent", Value: "codex"}},
		AuditPrompt: "You are Codex",
	}

	items := resolveEnabledItems([]string{"prompt-first", "header-last"})
	gotName, gotSource := detectBlacklistedUserAgent(items, signals)
	if gotName != "header-last" || gotSource != sourceRequestHeader {
		t.Fatalf("expected header stage to win before prompt stage, got name=%q source=%q", gotName, gotSource)
	}
}

func TestDetectBlacklistedUserAgent_StopsAtFirstMatchedItemInSameStage(t *testing.T) {
	restore := replaceItemsForTest(map[string]BlacklistItem{
		"first":  testPromptItem{name: "first", match: true},
		"second": testPromptItem{name: "second", match: true},
	})
	t.Cleanup(restore)

	items := resolveEnabledItems([]string{"first", "second"})
	gotName, gotSource := detectBlacklistedUserAgent(items, PreparedSignals{AuditPrompt: "matched"})
	if gotName != "first" || gotSource != sourceAuditPrompt {
		t.Fatalf("expected first prompt matcher to win, got name=%q source=%q", gotName, gotSource)
	}
}

func TestPrepareSignals_CollectsHeadersAuditPromptAndSystemMessageTexts(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutLogger(ctx, logrusx.New())
	sink := types.New("audit-1", logrusx.New())
	ctxhelper.PutAuditSink(ctx, sink)
	audithelper.Note(ctx, "prompt", "system prompt")

	req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/chat/completions", nil)
	req.Header.Set("User-Agent", "codex_cli_rs/0.116.0")
	ctxhelper.PutReverseProxyRequestInSnapshot(ctx, req)
	ctxhelper.PutMessageGroup(ctx, message.Group{
		RequestedMessages: message.Messages{
			openai.ChatCompletionMessage{Role: openai.ChatMessageRoleSystem, Content: "system 1"},
			openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: "user 1"},
		},
		AllMessages: message.Messages{
			openai.ChatCompletionMessage{Role: openai.ChatMessageRoleSystem, Content: "system 2"},
		},
	})

	signals := prepareSignals(ctx)
	if len(signals.HeaderPairs) != 1 || signals.HeaderPairs[0].Key != "User-Agent" || signals.HeaderPairs[0].Value != "codex_cli_rs/0.116.0" {
		t.Fatalf("unexpected header pairs: %#v", signals.HeaderPairs)
	}
	if signals.AuditPrompt != "system prompt" {
		t.Fatalf("expected audit prompt to be collected, got %q", signals.AuditPrompt)
	}
	if len(signals.MessageGroupTexts) != 2 || signals.MessageGroupTexts[0] != "system 1" || signals.MessageGroupTexts[1] != "system 2" {
		t.Fatalf("unexpected message-group texts: %#v", signals.MessageGroupTexts)
	}
}

type namedStubItem struct {
	name string
}

func (s namedStubItem) Name() string { return s.name }

type disabledStubItem struct {
	name string
}

func (s disabledStubItem) Name() string { return s.name }

func (s disabledStubItem) Enabled() bool { return false }

type testHeaderItem struct {
	name  string
	match bool
}

func (s testHeaderItem) Name() string { return s.name }

func (s testHeaderItem) MatchHeader(_, _ string) bool { return s.match }

type testPromptItem struct {
	name  string
	match bool
}

func (s testPromptItem) Name() string { return s.name }

func (s testPromptItem) MatchPrompt(string) bool { return s.match }

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
