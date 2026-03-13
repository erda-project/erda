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
	"testing"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/message"
)

func TestOpenClawItem_MatchMessageGroup(t *testing.T) {
	ctx := newDetectContextForTest()
	ctxhelper.PutMessageGroup(ctx, message.Group{
		RequestedMessages: message.Messages{
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: openClawSystemPromptHint,
			},
		},
	})

	matched, source := openClawItem{}.Match(ctx)
	if !matched || source != "message_group" {
		t.Fatalf("expected openclaw message-group match, got matched=%v source=%q", matched, source)
	}
}

func TestOpenClawItem_MatchAuditPrompt(t *testing.T) {
	ctx := newDetectContextForTest()
	ctxhelper.MustGetAuditSink(ctx).Note("prompt", "System: "+openClawSystemPromptHint)

	matched, source := openClawItem{}.Match(ctx)
	if !matched || source != "audit.prompt" {
		t.Fatalf("expected openclaw audit prompt match, got matched=%v source=%q", matched, source)
	}
}

func TestOpenClawItem_IgnoreOtherPrompt(t *testing.T) {
	ctx := newDetectContextForTest()
	ctxhelper.MustGetAuditSink(ctx).Note("prompt", "hello world")

	matched, source := openClawItem{}.Match(ctx)
	if matched || source != "" {
		t.Fatalf("expected prompt not to match openclaw, got matched=%v source=%q", matched, source)
	}
}

func newDetectContextForTest() context.Context {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutLogger(ctx, logrusx.New())
	ctxhelper.PutAuditSink(ctx, types.New("audit-1", logrusx.New()))
	return ctx
}
