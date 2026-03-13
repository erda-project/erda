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
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/message"
)

const openClawSystemPromptHint = "You are a personal assistant running inside OpenClaw"

type openClawItem struct{}

func init() {
	registerItem(openClawItem{})
}

func (openClawItem) Name() string {
	return "openclaw"
}

func (openClawItem) Match(ctx context.Context) (bool, string) {
	if msgGroup, ok := ctxhelper.GetMessageGroup(ctx); ok {
		if containsOpenClawInMessages(msgGroup.RequestedMessages) || containsOpenClawInMessages(msgGroup.AllMessages) {
			return true, "message_group"
		}
	}
	if sink, ok := ctxhelper.GetAuditSink(ctx); ok && sink != nil {
		if prompt, ok := sink.Snapshot()["prompt"]; ok && containsOpenClaw(asString(prompt)) {
			return true, "audit.prompt"
		}
	}
	return false, ""
}

func containsOpenClawInMessages(msgs message.Messages) bool {
	for _, msg := range msgs {
		if containsOpenClaw(chatMessageText(msg)) {
			return true
		}
	}
	return false
}

func containsOpenClaw(content string) bool {
	return strings.Contains(strings.ToLower(content), strings.ToLower(openClawSystemPromptHint))
}

func chatMessageText(msg openai.ChatCompletionMessage) string {
	if len(msg.MultiContent) == 0 {
		return msg.Content
	}
	parts := make([]string, 0, len(msg.MultiContent))
	for _, part := range msg.MultiContent {
		if part.Text != "" {
			parts = append(parts, part.Text)
		}
	}
	return strings.Join(parts, "\n")
}

func asString(value any) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", value)
}
