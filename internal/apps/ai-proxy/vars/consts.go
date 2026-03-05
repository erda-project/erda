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

package vars

const (
	XAIProxyHeaderPrefix       = "X-AI-Proxy-"
	XAIProxyChatType           = "X-AI-Proxy-ChatType"
	XAIProxyChatTitle          = "X-AI-Proxy-ChatTitle"
	XAIProxyChatId             = "X-AI-Proxy-ChatId"
	XAIProxySource             = "X-AI-Proxy-Source"
	XAIProxyName               = "X-AI-Proxy-Name"
	XAIProxyUsername           = "X-Ai-Proxy-Username"
	XAIProxyPhone              = "X-AI-Proxy-Phone"
	XAIProxyJobNumber          = "X-AI-Proxy-JobNumber"
	XAIProxyEmail              = "X-AI-Proxy-Email"
	XAIProxyDingTalkStaffID    = "X-AI-Proxy-DingTalkStaffID"
	XAIProxyProviderName       = "X-AI-Proxy-Provider-Name"
	XAIProxyProviderInstanceId = "X-AI-Proxy-Provider-Instance-Id"
	XAIProxyOrgId              = "X-Ai-Proxy-Org-Id"
	XAIProxyUserId             = "X-Ai-Proxy-User-Id"
	XAIProxyMetadata           = "X-Ai-Proxy-Metadata"
	XAiProxyErdaOpenapiSession = "X-Ai-Proxy-Erda-Openapi-Session"
	XRequestId                 = "X-Request-Id"
	XRequestIdLLMBackend       = "X-Request-Id-LLM-Backend"
	XAIProxyGeneratedCallId    = "X-AI-Proxy-Generated-Call-Id"

	XAIProxyModelId        = "X-AI-Proxy-Model-Id"
	XAIProxySessionId      = "X-AI-Proxy-Session-Id"
	XAIProxyPromptId       = "X-AI-Proxy-Prompt-Id"
	XAIProxyModel          = "X-AI-Proxy-Model"
	XAIProxyModelName      = "X-AI-Proxy-Model-Name"
	XAIProxyModelPublisher = "X-AI-Proxy-Model-Publisher"

	XAIProxyRequestBodyTransform     = "X-AI-Proxy-Request-Body-Transform"
	XAIProxyRequestThinkingTransform = "X-AI-Proxy-Request-Thinking-Transform"
	XAIProxyPolicyGroupTrace         = "X-AI-Proxy-Policy-Group-Trace"

	XAIProxyModelHealthMeta          = "X-AI-Proxy-Model-Health-Meta"
	XAIProxyModelHealthProbe         = "X-AI-Proxy-Model-Health-Probe"
	XAIProxyModelHealthMarkUnhealthy = "X-AI-Proxy-Model-Health-Mark-Unhealthy"
	XAIProxyModelRetryMeta           = "X-AI-Proxy-Model-Retry-Meta"

	XAIProxyForwardDialTimeout         = "X-AI-Proxy-Forward-Dial-Timeout"
	XAIProxyForwardTLSHandshakeTimeout = "X-AI-Proxy-Forward-TLS-Handshake-Timeout"
	XAIProxyForwardResponseTimeout     = "X-AI-Proxy-Forward-Response-Timeout"

	// XAIProxyRetry controls server-side transparent retries.
	//
	// - Default: enabled.
	// - Disable per request: set to "false".
	XAIProxyRetry = "X-AI-Proxy-Retry"
	// XAIProxyRetryDisabled explicitly disables server-side transparent retries
	// when set to a truthy bool value (for example: "true").
	XAIProxyRetryDisabled = "X-AI-Proxy-Retry-Disabled"
	// XAIProxyRetryMax overrides max attempt count (including first attempt),
	// e.g. "3" means first attempt + up to 2 retries.
	XAIProxyRetryMax = "X-AI-Proxy-Retry-Max"
	// IdempotencyKey is forwarded to upstream using call-id to deduplicate retries.
	IdempotencyKey = "Idempotency-Key"

	UIValueUndefined = "undefined"
)

const (
	EnvAIProxyAdminAuthKey = "AI_PROXY_ADMIN_AUTH_KEY"
)

const (
	McpScopeTypePlatform = "platform"
	McpScopeTypeClientId = "client"

	McpDefaultScopeType = McpScopeTypePlatform
	McpDefaultScopeId   = "0"
	McpAnyScopeType     = "*"
	McpAnyScopeId       = "*"
)
