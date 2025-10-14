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

package proxy_types

import (
	"context"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

type ProxyType string

const (
	// ProxyTypeUnknown means unknown proxy type
	ProxyTypeUnknown ProxyType = "UNKNOWN"

	// ProxyTypeOpenAI means openai style api
	ProxyTypeOpenAI ProxyType = "OPENAI"

	// ProxyTypeProxyBailian means proxy all api to bailian
	ProxyTypeProxyBailian ProxyType = "PROXY_BAILIAN"
	// ProxyTypeProxyBedrock means proxy all api to bedrock
	ProxyTypeProxyBedrock ProxyType = "PROXY_BEDROCK"
)

func DetermineProxyType(ctx context.Context) ProxyType {
	req, ok := ctxhelper.GetReverseProxyRequestInSnapshot(ctx)
	if !ok {
		return ProxyTypeUnknown
	}
	path := req.URL.Path
	if !strings.HasPrefix(path, "/proxy/") {
		return ProxyTypeOpenAI
	}
	switch {
	case strings.HasPrefix(path, "/proxy/bailian"):
		return ProxyTypeProxyBailian
	case strings.HasPrefix(path, "/proxy/bedrock"):
		return ProxyTypeProxyBedrock
	default:
		return ProxyTypeUnknown
	}
}
