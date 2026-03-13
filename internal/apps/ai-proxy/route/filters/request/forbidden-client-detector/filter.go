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
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/message"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
)

const (
	Name            = "forbidden-client-detector"
	envKeyBlacklist = "AI_PROXY_FORBIDDEN_CLIENT_BLACKLIST"

	clientOpenClaw           = "openclaw"
	openClawSystemPromptHint = "You are a personal assistant running inside OpenClaw"
)

type Config struct {
	Blacklist []string `json:"blacklist" yaml:"blacklist"`
}

type Filter struct {
	blacklist map[string]struct{}
}

var Creator filter_define.RequestRewriterCreator = func(_ string, config json.RawMessage) filter_define.ProxyRequestRewriter {
	var cfg Config
	if len(config) > 0 {
		if err := yaml.Unmarshal(config, &cfg); err != nil {
			panic(fmt.Errorf("failed to unmarshal forbidden client detector config: %w", err))
		}
	}
	cfg.Blacklist = resolveBlacklist(cfg.Blacklist)
	return &Filter{blacklist: toSet(cfg.Blacklist)}
}

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
}

func (f *Filter) OnProxyRequest(pr *httputil.ProxyRequest) error {
	ctx := pr.Out.Context()
	if !auth.IsClientToken(ctx) || len(f.blacklist) == 0 {
		return nil
	}

	clientName, source := f.detectForbiddenClient(ctx)
	if clientName == "" {
		return nil
	}
	if _, blocked := f.blacklist[clientName]; !blocked {
		return nil
	}

	audithelper.Note(ctx, "forbidden_client", clientName)
	audithelper.Note(ctx, "forbidden_client_match_source", source)
	return http_error.NewHTTPError(ctx, http.StatusForbidden, fmt.Sprintf("client token is not allowed for client: %s", clientName))
}

func (f *Filter) detectForbiddenClient(ctx context.Context) (string, string) {
	if msgGroup, ok := ctxhelper.GetMessageGroup(ctx); ok {
		if containsOpenClawInMessages(msgGroup.RequestedMessages) || containsOpenClawInMessages(msgGroup.AllMessages) {
			return clientOpenClaw, "message_group"
		}
	}
	if sink, ok := ctxhelper.GetAuditSink(ctx); ok && sink != nil {
		if prompt, ok := sink.Snapshot()["prompt"]; ok && containsOpenClaw(asString(prompt)) {
			return clientOpenClaw, "audit.prompt"
		}
	}
	return "", ""
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
	var parts []string
	for _, part := range msg.MultiContent {
		if part.Text != "" {
			parts = append(parts, part.Text)
		}
	}
	return strings.Join(parts, "\n")
}

func resolveBlacklist(configValues []string) []string {
	if raw := strings.TrimSpace(os.Getenv(envKeyBlacklist)); raw != "" {
		return splitAndNormalize(raw)
	}
	var values []string
	for _, item := range configValues {
		if normalized := normalize(item); normalized != "" {
			values = append(values, normalized)
		}
	}
	return values
}

func splitAndNormalize(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == ' ' || r == '\n' || r == '\t'
	})
	var values []string
	for _, part := range parts {
		if normalized := normalize(part); normalized != "" {
			values = append(values, normalized)
		}
	}
	return values
}

func normalize(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}

func toSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		result[value] = struct{}{}
	}
	return result
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
