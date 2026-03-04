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
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

var internalProbeToken = generateInternalProbeToken()

func IsHealthProbeRequest(headers http.Header) bool {
	if len(headers) == 0 {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(headers.Get(vars.XAIProxyHealthProbe)), "true")
}

func IsTrustedHealthProbeRequest(headers http.Header) bool {
	if !IsHealthProbeRequest(headers) {
		return false
	}
	token := strings.TrimSpace(headers.Get(vars.XAIProxyHealthProbeToken))
	if token == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(internalProbeToken)) == 1
}

func BuildProbeHeaders(headers http.Header) http.Header {
	cloned := cloneHeaders(headers)
	cloned.Set(vars.XAIProxyHealthProbe, "true")
	cloned.Set(vars.XAIProxyHealthProbeToken, internalProbeToken)
	return cloned
}

func generateInternalProbeToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("failed to generate internal health probe token: " + err.Error())
	}
	return hex.EncodeToString(b)
}
