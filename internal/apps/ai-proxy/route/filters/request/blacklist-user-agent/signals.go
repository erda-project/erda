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
	"sort"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/message"
)

type HeaderPair struct {
	Key   string
	Value string
}

type PreparedSignals struct {
	HeaderPairs       []HeaderPair
	AuditPrompt       string
	MessageGroupTexts []string
}

func prepareSignals(ctx context.Context) PreparedSignals {
	return PreparedSignals{
		HeaderPairs:       collectHeaderPairs(ctx),
		AuditPrompt:       collectAuditPrompt(ctx),
		MessageGroupTexts: collectMessageGroupTexts(ctx),
	}
}

func collectHeaderPairs(ctx context.Context) []HeaderPair {
	req, ok := ctxhelper.GetReverseProxyRequestInSnapshot(ctx)
	if !ok || req == nil {
		return nil
	}
	return flattenHeaders(req.Header)
}

func flattenHeaders(headers http.Header) []HeaderPair {
	if len(headers) == 0 {
		return nil
	}
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	pairs := make([]HeaderPair, 0, len(headers))
	for _, key := range keys {
		for _, value := range headers.Values(key) {
			pairs = append(pairs, HeaderPair{Key: key, Value: value})
		}
	}
	return pairs
}

func collectAuditPrompt(ctx context.Context) string {
	sink, ok := ctxhelper.GetAuditSink(ctx)
	if !ok || sink == nil {
		return ""
	}
	prompt, _ := sink.Snapshot()["prompt"].(string)
	return prompt
}

func collectMessageGroupTexts(ctx context.Context) []string {
	group, ok := ctxhelper.GetMessageGroup(ctx)
	if !ok {
		return nil
	}
	texts := make([]string, 0)
	texts = appendSystemMessageTexts(texts, group.RequestedMessages)
	texts = appendSystemMessageTexts(texts, group.AllMessages)
	return dedupTexts(texts)
}

func appendSystemMessageTexts(texts []string, msgs message.Messages) []string {
	for _, msg := range msgs {
		if msg.Role != openai.ChatMessageRoleSystem {
			continue
		}
		if text := chatMessageText(msg); text != "" {
			texts = append(texts, text)
		}
	}
	return texts
}

func dedupTexts(texts []string) []string {
	if len(texts) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(texts))
	deduped := make([]string, 0, len(texts))
	for _, text := range texts {
		if _, ok := seen[text]; ok {
			continue
		}
		seen[text] = struct{}{}
		deduped = append(deduped, text)
	}
	return deduped
}
