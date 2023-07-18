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
	XErdaAIProxySessionId       = "X-Erda-AI-Proxy-SessionId"
	XErdaAIProxyChatType        = "X-Erda-AI-Proxy-ChatType"
	XErdaAIProxyChatTitle       = "X-Erda-AI-Proxy-ChatTitle"
	XErdaAIProxyChatId          = "X-Erda-AI-Proxy-ChatId"
	XErdaAIProxySource          = "X-Erda-AI-Proxy-Source"
	XErdaAIProxyName            = "X-Erda-AI-Proxy-Name"
	XErdaAIProxyPhone           = "X-Erda-AI-Proxy-Phone"
	XErdaAIProxyJobNumber       = "X-Erda-AI-Proxy-JobNumber"
	XErdaAIProxyEmail           = "X-Erda-AI-Proxy-Email"
	XErdaAIProxyDingTalkStaffID = "X-Erda-AI-Proxy-DingTalkStaffID"
	XErdaAIProxyPrompt          = "X-Erda-AI-Proxy-Prompt"
	XAIProxyProvider            = "X-AI-Proxy-Provider"
	XAIProxyProviderInstance    = "X-AI-Proxy-Provider-Instance"
)

type (
	CtxKeyOrgSvc     struct{ CtxKeyOrgServer any }
	CtxKeyDAO        struct{ CtxKeyDatabaseAccess any }
	CtxKeyProviders  struct{ CtxKeyProviders any }
	MapKeyProvider   struct{ CtxKeyProvider any }
	MapKeyCredential struct{ CtxKeyCredential any }
)
