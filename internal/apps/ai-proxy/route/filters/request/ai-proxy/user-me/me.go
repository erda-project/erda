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

package user_me

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"strconv"

	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth/akutil"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/transports"
	httperrorutil "github.com/erda-project/erda/pkg/http/httputil"
)

var _ filter_define.ProxyRequestRewriter = (*Filter)(nil)
var Creator filter_define.RequestRewriterCreator = func(_ string, _ json.RawMessage) filter_define.ProxyRequestRewriter { return &Filter{} }

func init() { filter_define.RegisterFilterCreator("user-me", Creator) }

type Filter struct{}

type Me struct {
	IsAdmin       bool             `json:"is_admin"`
	IsClientToken bool             `json:"is_client_token"`
	IsClient      bool             `json:"is_client"`
	Client        *clientpb.Client `json:"client,omitempty"`
}

func (f *Filter) OnProxyRequest(pr *httputil.ProxyRequest) error {
	ctx := pr.In.Context()

	var me Me

	// check admin key first
	isAdmin, err := akutil.CheckAdmin(ctx, pr.In, ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager))
	if err != nil {
		return http_error.NewHTTPError(ctx, http.StatusUnauthorized, err.Error())
	}
	me.IsAdmin = isAdmin
	if !isAdmin {
		// check client info
		clientToken, client, err := akutil.CheckAkOrToken(ctx, pr.In, ctxhelper.MustGetCacheManager(ctx).(cachetypes.Manager))
		if err != nil {
			return http_error.NewHTTPError(ctx, http.StatusUnauthorized, err.Error())
		}
		me.IsClientToken = clientToken != nil
		me.IsClient = clientToken == nil && client != nil
		if me.IsClient && client != nil {
			me.Client = &clientpb.Client{
				Name:      client.Name,
				Desc:      client.Desc,
				CreatedAt: client.CreatedAt,
			}
		}
	}

	meBytes, err := json.Marshal(me)
	if err != nil {
		return http_error.NewHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}

	// construct response
	respHeader := http.Header{}
	respHeader.Set(httperrorutil.HeaderKeyContentType, string(httperrorutil.ApplicationJson))
	respHeader.Set(httperrorutil.HeaderKeyContentLength, strconv.FormatInt(int64(len(meBytes)), 10))

	resp := &http.Response{
		StatusCode:    http.StatusOK,
		Header:        respHeader,
		Body:          io.NopCloser(bytes.NewReader(meBytes)),
		ContentLength: int64(len(meBytes)),
		Request:       pr.Out,
	}

	// trigger filter-generated response
	transports.TriggerRequestFilterGeneratedResponse(pr.Out, resp)

	return nil
}
