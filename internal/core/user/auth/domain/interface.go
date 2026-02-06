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

	identitypb "github.com/erda-project/erda-proto-go/core/user/identity/pb"
	"github.com/erda-project/erda/internal/core/user/common"
)

type RefreshWriter interface {
	WriteRefresh(rw http.ResponseWriter, req *http.Request, refresh *identitypb.SessionRefresh) error
}

// SessionRevoker (Optional: revoke session, e.g. remove from redis)
// TODO:
type SessionRevoker interface {
	Revoke(ctx context.Context) error
}

type UserAuthFacade interface {
	NewState() UserAuthState
}

type UserAuthState interface {
	GetOrgInfo(orgHeader, domainHeader string) (orgID uint64, err error)
	IsLogin(req *http.Request) UserAuthResult
	GetInfo(req *http.Request) (*common.UserInfo, UserAuthResult)
	GetScopeInfo(req *http.Request) (common.UserScopeInfo, UserAuthResult)
	Login(code string, queryValues url.Values) error
	PwdLogin(username, password string) error
}
