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
	{BackendPath: "/api/ai-proxy/access", Method: http.MethodPost},

	{BackendPath: "/api/ai-proxy/chat-logs", Method: http.MethodGet},
	{BackendPath: "/api/ai-proxy/sessions/{sessionId}/chat-logs", Method: http.MethodGet},

	{BackendPath: "/api/ai-proxy/credentials", Method: http.MethodPost},
	{BackendPath: "/api/ai-proxy/credentials", Method: http.MethodGet},
	{BackendPath: "/api/ai-proxy/credentials/{accessKeyId}", Method: http.MethodDelete},
	{BackendPath: "/api/ai-proxy/credentials/{accessKeyId}", Method: http.MethodPut},
	{BackendPath: "/api/ai-proxy/credentials/{accessKeyId}", Method: http.MethodGet},

	{BackendPath: "/api/ai-proxy/models", Method: http.MethodGet},

	{BackendPath: "/api/ai-proxy/providers", Method: http.MethodPost},
	{BackendPath: "/api/ai-proxy/providers", Method: http.MethodGet},
	{BackendPath: "/api/ai-proxy/providers/{name}/instances/{instanceId}", Method: http.MethodDelete},
	{BackendPath: "/api/ai-proxy/providers/{name}/instances/{instanceId}", Method: http.MethodPut},
	{BackendPath: "/api/ai-proxy/providers/{name}/instances/{instanceId}", Method: http.MethodGet},

	{BackendPath: "/api/ai-proxy/sessions", Method: http.MethodPost},
	{BackendPath: "/api/ai-proxy/sessions", Method: http.MethodGet},
	{BackendPath: "/api/ai-proxy/sessions/{id}", Method: http.MethodDelete},
	{BackendPath: "/api/ai-proxy/sessions/{id}", Method: http.MethodPut},
	{BackendPath: "/api/ai-proxy/sessions/{id}", Method: http.MethodGet},
	{BackendPath: "/api/ai-proxy/sessions/{id}/actions/reset", Method: http.MethodPatch},
	{BackendPath: "/api/ai-proxy/sessions/{id}/actions/archive", Method: http.MethodPatch},
}
