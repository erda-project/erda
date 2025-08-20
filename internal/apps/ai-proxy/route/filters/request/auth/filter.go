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

package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"strings"

	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	clienttokenpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_token/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/client_token"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	httperrorutil "github.com/erda-project/erda/pkg/http/httputil"
)

const Name = "auth"

var (
	_ filter_define.ProxyRequestRewriter = (*Context)(nil)
)

type Context struct {
}

var ContextCreator filter_define.RequestRewriterCreator = func(name string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &Context{}
}

func init() {
	filter_define.RegisterFilterCreator(Name, ContextCreator)
}

func (f *Context) OnProxyRequest(pr *httputil.ProxyRequest) error {
	ctx := pr.Out.Context()

	var (
		l = ctxhelper.MustGetLogger(ctx)
		q = ctxhelper.MustGetDBClient(ctx)
	)

	// find client
	var client *clientpb.Client
	ak := vars.TrimBearer(pr.Out.Header.Get(httperrorutil.HeaderKeyAuthorization))
	if ak == "" {
		return http_error.NewHTTPError(http.StatusUnauthorized, "Authorization is required")
	}
	if strings.HasPrefix(ak, client_token.TokenPrefix) {
		tokenPagingResp, err := q.ClientTokenClient().Paging(ctx, &clienttokenpb.ClientTokenPagingRequest{
			PageSize: 1,
			PageNum:  1,
			Token:    ak,
		})
		if err != nil || tokenPagingResp.Total < 1 {
			l.Errorf("failed to get client token, token: %s, err: %v", ak, err)
			return http_error.NewHTTPError(http.StatusForbidden, "failed to get client token")
		}
		token := tokenPagingResp.List[0]
		clientResp, err := q.ClientClient().Get(ctx, &clientpb.ClientGetRequest{ClientId: token.ClientId})
		if err != nil {
			l.Errorf("failed to get client, id: %s, err: %v", tokenPagingResp.List[0].ClientId, err)
			return http_error.NewHTTPError(http.StatusForbidden, "Authorization is invalid")
		}
		client = clientResp
		ctxhelper.PutClientToken(ctx, token)
	} else {
		clientPagingResult, err := q.ClientClient().Paging(ctx, &clientpb.ClientPagingRequest{
			AccessKeyIds: []string{ak},
			PageNum:      1,
			PageSize:     1,
		})
		if err != nil || clientPagingResult.Total < 1 {
			l.Errorf("failed to get client, access_key_id: %s, err: %v", ak, err)
			return http_error.NewHTTPError(http.StatusForbidden, "Authorization is invalid")
		}
		client = clientPagingResult.List[0]
	}

	// store data to context
	ctxhelper.PutClient(ctx, client)

	return nil
}
