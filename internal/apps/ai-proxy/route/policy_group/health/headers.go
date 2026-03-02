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

package health

import (
	"net/http"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

var probeForwardedHeaders = []string{
	"Authorization",
	"AK-Token",
	"X-AK-Token",
	"X-Access-Key",
	"X-API-Key",
	"Api-Key",
	"Cookie",
}

func IsHealthProbeRequest(headers http.Header) bool {
	if len(headers) == 0 {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(headers.Get(vars.XAIProxyHealthProbe)), "true")
}

func BuildProbeHeaders(headers http.Header) http.Header {
	if len(headers) == 0 {
		return http.Header{}
	}
	out := make(http.Header, len(probeForwardedHeaders))
	for _, key := range probeForwardedHeaders {
		values := headers.Values(key)
		for _, value := range values {
			out.Add(key, value)
		}
	}
	return out
}
