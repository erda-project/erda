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
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
)

const Name = "blacklist-user-agent"

const (
	sourceRequestHeader = "request_header"
	sourceAuditPrompt   = "audit.prompt"
	sourceMessageGroup  = "message_group"
)

type Filter struct{}

var (
	Creator = filter_define.RequestRewriterCreator(func(_ string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
		return &Filter{}
	})
)

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
}

func (f *Filter) OnProxyRequest(pr *httputil.ProxyRequest) error {
	ctx := pr.Out.Context()
	enabledItems := resolveEnabledItems(getBlacklistByCredential(ctx))
	if len(enabledItems) == 0 {
		return nil
	}
	signals := prepareSignals(ctx)

	userAgentName, source := detectBlacklistedUserAgent(enabledItems, signals)
	if userAgentName == "" {
		return nil
	}

	audithelper.Note(ctx, "blacklist_user_agent", userAgentName)
	audithelper.Note(ctx, "blacklist_user_agent_match_source", source)
	return http_error.NewHTTPError(ctx, http.StatusForbidden, fmt.Sprintf("request is not allowed for blacklisted user-agent: %s", userAgentName))
}

func detectBlacklistedUserAgent(items []BlacklistItem, signals PreparedSignals) (string, string) {
	if itemName := detectBlacklistedUserAgentFromHeaders(items, signals.HeaderPairs); itemName != "" {
		return itemName, sourceRequestHeader
	}
	if itemName := detectBlacklistedUserAgentFromPrompt(items, signals.AuditPrompt); itemName != "" {
		return itemName, sourceAuditPrompt
	}
	if itemName := detectBlacklistedUserAgentFromMessageGroup(items, signals.MessageGroupTexts); itemName != "" {
		return itemName, sourceMessageGroup
	}
	return "", ""
}

func detectBlacklistedUserAgentFromHeaders(items []BlacklistItem, headerPairs []HeaderPair) string {
	if len(headerPairs) == 0 {
		return ""
	}
	for _, item := range items {
		matcher, ok := item.(HeaderMatcher)
		if !ok {
			continue
		}
		for _, pair := range headerPairs {
			if matcher.MatchHeader(pair.Key, pair.Value) {
				return item.Name()
			}
		}
	}
	return ""
}

func detectBlacklistedUserAgentFromPrompt(items []BlacklistItem, prompt string) string {
	if prompt == "" {
		return ""
	}
	for _, item := range items {
		matcher, ok := item.(PromptMatcher)
		if ok && matcher.MatchPrompt(prompt) {
			return item.Name()
		}
	}
	return ""
}

func detectBlacklistedUserAgentFromMessageGroup(items []BlacklistItem, texts []string) string {
	if len(texts) == 0 {
		return ""
	}
	for _, item := range items {
		matcher, ok := item.(MessageGroupMatcher)
		if !ok {
			continue
		}
		for _, text := range texts {
			if matcher.MatchMessageGroupText(text) {
				return item.Name()
			}
		}
	}
	return ""
}

func getBlacklistByCredential(ctx context.Context) []string {
	cfg := getConfig()
	switch {
	case auth.IsClientToken(ctx):
		return cfg.ClientToken.Blacklist
	case auth.IsClient(ctx):
		return cfg.Client.Blacklist
	default:
		return nil
	}
}
