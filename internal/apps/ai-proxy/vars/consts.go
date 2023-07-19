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
	XAIProxySessionId          = "X-AI-Proxy-SessionId"
	XAIProxyChatType           = "X-AI-Proxy-ChatType"
	XAIProxyChatTitle          = "X-AI-Proxy-ChatTitle"
	XAIProxyChatId             = "X-AI-Proxy-ChatId"
	XAIProxySource             = "X-AI-Proxy-Source"
	XAIProxyName               = "X-AI-Proxy-Name"
	XAIProxyPhone              = "X-AI-Proxy-Phone"
	XAIProxyJobNumber          = "X-AI-Proxy-JobNumber"
	XAIProxyEmail              = "X-AI-Proxy-Email"
	XAIProxyDingTalkStaffID    = "X-AI-Proxy-DingTalkStaffID"
	XAIProxyPrompt             = "X-AI-Proxy-Prompt"
	XAIProxyProviderId         = "X-AI-Proxy-Provider-Name"
	XAIProxyProviderInstanceId = "X-AI-Proxy-Provider-Instance-Id"
	XAIProxyOrgId              = "X-Ai-Proxy-Org-Id"
	XAIProxyUserId             = "X-Ai-Proxy-UserId"
)

type (
	CtxKeyOrgSvc      struct{ CtxKeyOrgServer any }
	CtxKeyDAO         struct{ CtxKeyDatabaseAccess any }
	CtxKeyProviders   struct{ CtxKeyProviders any }
	MapKeyProvider    struct{ CtxKeyProvider any }
	MapKeyCredential  struct{ CtxKeyCredential any }
	CtxKeyErdaOpenapi struct{ ErdaOpenapi any }
)
