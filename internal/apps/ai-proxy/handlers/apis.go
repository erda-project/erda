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

package handlers

import (
	"net/http"

	dynamic "github.com/erda-project/erda-proto-go/core/openapi/dynamic-register/pb"
)

var APIs = []*dynamic.API{
	{Path: "/api/ai-proxy/access", BackendPath: "/access", Method: http.MethodPost},

	{Path: "/api/ai-proxy/chat-logs", BackendPath: "chat-logs", Method: http.MethodGet},
	{Path: "/api/ai-proxy/sessions/{sessionId}/chat-logs", BackendPath: "/sessions/{sessionId}/chat-logs", Method: http.MethodGet},

	{Path: "/api/ai-proxy/credentials", BackendPath: "/credentials", Method: http.MethodPost},
	{Path: "/api/ai-proxy/credentials", BackendPath: "/credentials", Method: http.MethodGet},
	{Path: "/api/ai-proxy/credentials/{accessKeyId}", BackendPath: "/credentials/{accessKeyId}", Method: http.MethodDelete},
	{Path: "/api/ai-proxy/credentials/{accessKeyId}", BackendPath: "/credentials/{accessKeyId}", Method: http.MethodPut},
	{Path: "/api/ai-proxy/credentials/{accessKeyId}", BackendPath: "/credentials/{accessKeyId}", Method: http.MethodGet},

	{Path: "/api/ai-proxy/models", BackendPath: "/models", Method: http.MethodGet},

	{Path: "/api/ai-proxy/providers", BackendPath: "/providers", Method: http.MethodPost},
	{Path: "/api/ai-proxy/providers", BackendPath: "/providers", Method: http.MethodGet},
	{Path: "/api/ai-proxy/providers/{name}/instances/{instanceId}", BackendPath: "/providers/{name}/instances/{instanceId}", Method: http.MethodDelete},
	{Path: "/api/ai-proxy/providers/{name}/instances/{instanceId}", BackendPath: "/providers/{name}/instances/{instanceId}", Method: http.MethodPut},
	{Path: "/api/ai-proxy/providers/{name}/instances/{instanceId}", BackendPath: "/providers/{name}/instances/{instanceId}", Method: http.MethodGet},

	{Path: "/api/ai-proxy/sessions", BackendPath: "/sessions", Method: http.MethodPost},
	{Path: "/api/ai-proxy/sessions", BackendPath: "/sessions", Method: http.MethodGet},
	{Path: "/api/ai-proxy/sessions/{id}", BackendPath: "/sessions/{id}", Method: http.MethodDelete},
	{Path: "/api/ai-proxy/sessions/{id}", BackendPath: "/sessions/{id}", Method: http.MethodPut},
	{Path: "/api/ai-proxy/sessions/{id}", BackendPath: "/sessions/{id}", Method: http.MethodGet},
	{Path: "/api/ai-proxy/sessions/{id}/actions/reset", BackendPath: "/sessions/{id}/actions/reset", Method: http.MethodPatch},
	{Path: "/api/ai-proxy/sessions/{id}/actions/archive", BackendPath: "/sessions/{id}/actions/archive", Method: http.MethodPatch},
}
