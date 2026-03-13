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
	"encoding/json"
	"strings"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/message"
)

// Source: https://github.com/search?q=repo:openclaw/openclaw+%22You+are+a+personal+assistant+running+inside+OpenClaw.%22&type=code
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
		if containsOpenClawSystemMessage(msgGroup.RequestedMessages) || containsOpenClawSystemMessage(msgGroup.AllMessages) {
			return true, "message_group"
		}
	}
	if bodyValue, ok := ctxhelper.GetReverseProxyRequestBodyBytes(ctx); ok {
		if matched, source := matchOpenClawFromRequestBody(bodyValue); matched {
			return true, source
		}
	}
	return false, ""
}

func containsOpenClawSystemMessage(msgs message.Messages) bool {
	for _, msg := range msgs {
		if msg.Role != openai.ChatMessageRoleSystem {
			continue
		}
		if isOpenClawSystemPrompt(chatMessageText(msg)) {
			return true
		}
	}
	return false
}

func isOpenClawSystemPrompt(content string) bool {
	return strings.TrimSpace(content) == openClawSystemPromptHint
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

func matchOpenClawFromRequestBody(value any) (bool, string) {
	bodyBytes, ok := value.([]byte)
	if !ok || len(bodyBytes) == 0 {
		return false, ""
	}

	var req struct {
		Messages     []openai.ChatCompletionMessage `json:"messages"`
		Instructions string                         `json:"instructions"`
	}
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		return false, ""
	}
	if containsOpenClawSystemMessage(req.Messages) {
		return true, "request_body.messages"
	}
	if isOpenClawSystemPrompt(req.Instructions) {
		return true, "request_body.instructions"
	}
	return false, ""
}
