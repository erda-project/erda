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

package blacklist_user_agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
)

const Name = "blacklist-user-agent"

type Filter struct{}

var (
	Creator = filter_define.RequestRewriterCreator(func(_ string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
		return &Filter{}
	})
)

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
}

func (f *Filter) OnProxyRequest(pr *httputil.ProxyRequest) error {
	ctx := pr.Out.Context()
	if !auth.IsClientToken(ctx) {
		return nil
	}

	userAgentName, source := detectBlacklistedUserAgent(ctx)
	if userAgentName == "" {
		return nil
	}

	audithelper.Note(ctx, "blacklist_user_agent", userAgentName)
	audithelper.Note(ctx, "blacklist_user_agent_match_source", source)
	return http_error.NewHTTPError(ctx, http.StatusForbidden, fmt.Sprintf("client token is not allowed for blacklisted user-agent: %s", userAgentName))
}

func detectBlacklistedUserAgent(ctx context.Context) (string, string) {
	cfg := getConfig()
	for _, itemName := range cfg.Blacklist {
		item, ok := getItem(itemName)
		if !ok {
			continue
		}
		if matched, source := item.Match(ctx); matched {
			return item.Name(), source
		}
	}
	return "", ""
}
