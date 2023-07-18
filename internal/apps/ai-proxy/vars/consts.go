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
	XErdaAIProxySessionId       = "X-AI-Proxy-SessionId"
	XErdaAIProxyChatType        = "X-AI-Proxy-ChatType"
	XErdaAIProxyChatTitle       = "X-AI-Proxy-ChatTitle"
	XErdaAIProxyChatId          = "X-AI-Proxy-ChatId"
	XErdaAIProxySource          = "X-AI-Proxy-Source"
	XErdaAIProxyName            = "X-AI-Proxy-Name"
	XErdaAIProxyPhone           = "X-AI-Proxy-Phone"
	XErdaAIProxyJobNumber       = "X-AI-Proxy-JobNumber"
	XErdaAIProxyEmail           = "X-AI-Proxy-Email"
	XErdaAIProxyDingTalkStaffID = "X-AI-Proxy-DingTalkStaffID"
	XErdaAIProxyPrompt          = "X-AI-Proxy-Prompt"
	XAIProxyProvider            = "X-AI-Proxy-Provider"
	XAIProxyProviderInstance    = "X-AI-Proxy-Provider-Instance"
	XAIProxyOrgId               = "X-Ai-Proxy-Org-Id"
	XAIProxyUserId              = "X-Ai-Proxy-UserId"
)

type (
	CtxKeyOrgSvc      struct{ CtxKeyOrgServer any }
	CtxKeyDAO         struct{ CtxKeyDatabaseAccess any }
	CtxKeyProviders   struct{ CtxKeyProviders any }
	MapKeyProvider    struct{ CtxKeyProvider any }
	MapKeyCredential  struct{ CtxKeyCredential any }
	CtxKeyErdaOpenapi struct{ ErdaOpenapi any }
)
