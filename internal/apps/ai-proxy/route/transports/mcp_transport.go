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

package transports

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/pkg/clusterdialer"
)

type McpTransport struct {
	http.RoundTripper
}

func NewMcpTransport() *McpTransport {
	return &McpTransport{
		BaseTransport,
	}
}

func (t *McpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport, ok := t.RoundTripper.(*http.Transport)
	if !ok {
		logrus.Infof("failed to cast transport to http.Transport")
		return t.RoundTripper.RoundTrip(req)
	}

	info, ok := ctxhelper.GetMcpInfo(req.Context())
	if !ok {
		logrus.Infof("failed to cast context to mcp info")
		return t.RoundTripper.RoundTrip(req)
	}

	if info.ClusterName != "" {
		logrus.Infof("with clusterdialer, cluster name: %s", info.ClusterName)
		transport.DialContext = clusterdialer.DialContext(info.ClusterName)
	}

	return transport.RoundTrip(req)
}
