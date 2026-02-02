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

package facade

import (
	"context"
	"net/http"

	identitypb "github.com/erda-project/erda-proto-go/core/user/identity/pb"
)

type PersistedCredential struct {
	Type        identitypb.TokenSource
	AccessToken string `json:"accessToken"`
	// CookieName is set when Type == Cookie (from CredentialStore). Used to build GetCurrentUserRequest.
	CookieName string
}

type CredentialStore interface {
	Load(ctx context.Context, req *http.Request) (*PersistedCredential, error)
}

type cookieStore struct {
	cookieName string
}

func (c *cookieStore) Load(_ context.Context, req *http.Request) (*PersistedCredential, error) {
	cookie, err := req.Cookie(c.cookieName)
	if err != nil {
		return nil, err
	}
	return &PersistedCredential{
		Type:        identitypb.TokenSource_Cookie,
		AccessToken: cookie.Value,
		CookieName:  c.cookieName,
	}, nil
}

func NewCookieStore(cookieName string) CredentialStore {
	return &cookieStore{
		cookieName: cookieName,
	}
}

// ToGetCurrentUserRequest builds identity GetCurrentUserRequest from credential. Used by facade callers (e.g. userState).
func ToGetCurrentUserRequest(cred *PersistedCredential) *identitypb.GetCurrentUserRequest {
	req := &identitypb.GetCurrentUserRequest{
		AccessToken: cred.AccessToken,
		Source:      cred.Type,
	}
	if cred.CookieName != "" {
		req.CookieName = &cred.CookieName
	}
	return req
}
