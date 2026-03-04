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
	"testing"

	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

func TestBuildProbeHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer t_x")
	headers.Set("AK-Token", "ak-token-x")
	headers.Set("X-Trace-Id", "trace-1")
	headers.Set(vars.XAIProxyForwardDialTimeout, "1ms")
	headers.Set(vars.XAIProxyForwardTLSHandshakeTimeout, "1ms")
	probeHeaders := BuildProbeHeaders(headers)

	if probeHeaders.Get("Authorization") == "" || probeHeaders.Get("AK-Token") == "" {
		t.Fatalf("expected headers kept, got: %v", probeHeaders)
	}
	if probeHeaders.Get("X-Trace-Id") != "trace-1" {
		t.Fatalf("expected x-trace-id kept, got: %v", probeHeaders)
	}
	if probeHeaders.Get(vars.XAIProxyModelHealthProbe) != "true" {
		t.Fatalf("expected probe marker kept, got: %v", probeHeaders)
	}
	if probeHeaders.Get(vars.XAIProxyForwardDialTimeout) != "" {
		t.Fatalf("expected forward dial timeout dropped, got: %v", probeHeaders)
	}
	if probeHeaders.Get(vars.XAIProxyForwardTLSHandshakeTimeout) != "" {
		t.Fatalf("expected forward tls handshake timeout dropped, got: %v", probeHeaders)
	}
}
