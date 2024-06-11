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

	XAIProxyModelId   = "X-AI-Proxy-Model-Id"
	XAIProxySessionId = "X-AI-Proxy-Session-Id"
	XAIProxyPromptId  = "X-AI-Proxy-Prompt-Id"

	UIValueUndefined = "undefined"
)

const (
	EnvAIProxyAdminAuthKey = "AI_PROXY_ADMIN_AUTH_KEY"
)

type (
	CtxKeyDAO         struct{ CtxKeyDAO any }
	CtxKeyErdaOpenapi struct{ CtxKeyErdaOpenapi any }
	CtxKeyIsAdmin     struct{ CtxKeyIsAdmin bool }
	CtxKeyClientId    struct{ CtxKeyClientId string }
	CtxKeyClient      struct{ CtxKeyClient any }

	CtxKeyRichClientHandler struct{ CtxKeyRichClientHandler any }

	MapKeyClient         struct{ MapKeyClient any }
	MapKeyModel          struct{ MapKeyModel any }
	MapKeyModelProvider  struct{ MapKeyModelProvider any }
	MapKeyPromptTemplate struct{ MapKeyPromptTemplate any }
	MapKeySession        struct{ MapKeySession any }
	MapKeyClientToken    struct{ MapKeyClientToken any }
	MapKeyMessageGroup   struct{ MapKeyMessageGroup any }
	MapKeyUserPrompt     struct{ MapKeyUserPrompt any }
	MapKeyIsStream       struct{ MapKeyIsStream any }
	MapKeyAudit          struct{ MapKeyAudit any }
	MapKeyAudioInfo      struct{ MapKeyAudioInfo any }
	MapKeyImageInfo      struct{ MapKeyImageInfo any }

	MapKeyLLMDirectorPassedOnRequest      struct{ MapKeyLLMDirectorPassedOnRequest any }
	MapKeyLLMDirectorActualResponseWriter struct{ MapKeyLLMDirectorActualResponseWriter any }
)
