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

package legacy

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	tokenpb "github.com/erda-project/erda-proto-go/core/token/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/openapi/legacy/auth"
	"github.com/erda-project/erda/internal/core/openapi/legacy/component-protocol/generate/auto_register"
	"github.com/erda-project/erda/internal/core/openapi/legacy/hooks"
	"github.com/erda-project/erda/internal/core/openapi/legacy/hooks/prehandle"
	"github.com/erda-project/erda/internal/core/openapi/settings"
	"github.com/erda-project/erda/pkg/oauth2"
	"github.com/erda-project/erda/pkg/strutil"
)

type LoginServer struct {
	r    http.Handler
	auth *auth.Auth

	oauth2server *oauth2.OAuth2Server
}

func NewLoginServer(token tokenpb.TokenServiceServer, settings settings.OpenapiSettings) (*LoginServer, error) {
	oauth2server := oauth2.NewOAuth2Server()
	auth, err := auth.NewAuth(oauth2server, token, settings)
	if err != nil {
		return nil, err
	}
	bdl := bundle.New(
		bundle.WithDOP(),
		bundle.WithPipeline(),
		bundle.WithErdaServer(),
		bundle.WithMonitor(),
		bundle.WithTMC(),
	)
	auto_register.RegisterAll()
	h, err := NewReverseProxyWithAuth(auth, bdl)
	if err != nil {
		return nil, err
	}
	return &LoginServer{r: h, auth: auth, oauth2server: oauth2server}, nil
}

func (s *LoginServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if hooks.Enable {
		prehandle.ReplaceOldCookie(context.Background(), rw, req)
	}
	//prehandle.FilterCookie(context.Background(), rw, req) // for auth
	if strutil.HasPrefixes(req.URL.Path, "/oauth2") {
		switch req.URL.Path {
		case "/oauth2/token":
			s.oauth2server.Token(rw, req)
		case "/oauth2/invalidate_token":
			s.oauth2server.InvalidateToken(rw, req)
		case "/oauth2/validate_token":
			s.oauth2server.ValidateToken(rw, req)
		default:
			errStr := fmt.Sprintf("not found path: %v", req.URL)
			logrus.Error(errStr)
			http.Error(rw, errStr, 404)
			return
		}
	} else {
		s.r.ServeHTTP(rw, req)
	}
}
