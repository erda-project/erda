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

package domain

import (
	"context"
	"net/http"
	"net/url"

	"github.com/erda-project/erda/internal/core/user/common"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type OAuthProvider interface {
	OAuthTokenProvider
	OAuthLoginFlow
}

type OAuthLoginFlow interface {
	AuthURL(ctx context.Context, referer string) (string, error)
}

type OAuthTokenProvider interface {
	ExchangeCode(ctx context.Context, code string, extraParams url.Values) (*OAuthToken, error)
	ExchangePassword(ctx context.Context, username, password string, extraParams url.Values) (*OAuthToken, error)
	ExchangeClientCredentials(ctx context.Context, refresh bool, extraParams url.Values) (*OAuthToken, error)
}

type RequestAuthenticator interface {
	Apply(req *httpclient.Request)
}

type Identity interface {
	Me(ctx context.Context, authCtx RequestAuthenticator) (*common.UserInfo, error)
}

type CredentialStore interface {
	Load(ctx context.Context, req *http.Request) (*PersistedCredential, error)
	Persist(ctx context.Context, cred *AuthCredential) (*PersistedCredential, error)
}

type RefreshWriter interface {
	WriteRefresh(rw http.ResponseWriter, req *http.Request, refresh *common.SessionRefresh) error
}

// SessionRevoker (Optional: revoke session, e.g. remove from redis)
type SessionRevoker interface {
	Revoke(ctx context.Context, sessionID string) error
}
