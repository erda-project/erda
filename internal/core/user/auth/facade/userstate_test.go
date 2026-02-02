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
	"testing"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	identitypb "github.com/erda-project/erda-proto-go/core/user/identity/pb"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
	"github.com/erda-project/erda/internal/core/user/common"
)

// mockCredStore fails Load or returns fixed credential.
type mockCredStore struct {
	cred *PersistedCredential
	err  error
}

func (m *mockCredStore) Load(_ context.Context, _ *http.Request) (*PersistedCredential, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.cred, nil
}

// mockIdentitySvc returns fixed GetCurrentUser response or error.
type mockIdentitySvc struct {
	identitypb.UnimplementedUserIdentityServiceServer
	resp *identitypb.GetCurrentUserResponse
	err  error
}

func (m *mockIdentitySvc) GetCurrentUser(_ context.Context, _ *identitypb.GetCurrentUserRequest) (*identitypb.GetCurrentUserResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.resp, nil
}

func TestUserState_IsLogin_NoCredential(t *testing.T) {
	u := &userState{
		state:       GetInit,
		credStore:   &mockCredStore{err: http.ErrNoCookie},
		IdentitySvc: &mockIdentitySvc{},
	}
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	result := u.IsLogin(req)
	if result.Code != domain.Unauthed {
		t.Errorf("IsLogin without credential expected Unauthed, got code %d detail %q", result.Code, result.Detail)
	}
}

func TestUserState_IsLogin_WithCredential(t *testing.T) {
	u := &userState{
		state:       GetInit,
		credStore:   &mockCredStore{cred: &PersistedCredential{Type: identitypb.TokenSource_Cookie, AccessToken: "at", CookieName: "sid"}},
		IdentitySvc: &mockIdentitySvc{}, // not called when targetState is GotToken
	}
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	result := u.IsLogin(req)
	if result.Code != domain.AuthSuccess {
		t.Errorf("IsLogin with credential expected AuthSuccess, got code %d detail %q", result.Code, result.Detail)
	}
}

func TestUserState_GetInfo_IdentityFails(t *testing.T) {
	u := &userState{
		state:       GetInit,
		credStore:   &mockCredStore{cred: &PersistedCredential{AccessToken: "at"}},
		IdentitySvc: &mockIdentitySvc{err: context.DeadlineExceeded},
	}
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.Background())
	info, result := u.GetInfo(req)
	if result.Code != domain.Unauthed {
		t.Errorf("GetInfo when Identity fails expected Unauthed, got code %d", result.Code)
	}
	if info != nil {
		t.Error("GetInfo on error should return nil info")
	}
}

func TestUserState_GetInfo_Success(t *testing.T) {
	u := &userState{
		state:     GetInit,
		credStore: &mockCredStore{cred: &PersistedCredential{AccessToken: "at"}},
		IdentitySvc: &mockIdentitySvc{
			resp: &identitypb.GetCurrentUserResponse{
				Data: &commonpb.UserInfo{Id: "100", Name: "testuser", Nick: "Test"},
			},
		},
		bundle: nil, // not used when targetState is GotInfo
	}
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.Background())
	info, result := u.GetInfo(req)
	if result.Code != domain.AuthSuccess {
		t.Fatalf("GetInfo expected AuthSuccess, got code %d detail %q", result.Code, result.Detail)
	}
	if info == nil {
		t.Fatal("GetInfo expected non-nil user info")
	}
	if info.Id != "100" || info.Name != "testuser" {
		t.Errorf("GetInfo user: got Id=%q Name=%q", info.Id, info.Name)
	}
}

func TestUserState_GetScopeInfo_NoOrgHeader(t *testing.T) {
	// GetScopeInfo runs get(req, GotScopeInfo). From GotInfo we need orgHeader/domainHeader to get org;
	// when both empty GetOrgInfo returns 0, we skip bundle.ScopeRoleAccess and set scopeInfo with OrgID 0.
	u := &userState{
		state:     GetInit,
		credStore: &mockCredStore{cred: &PersistedCredential{AccessToken: "at"}},
		IdentitySvc: &mockIdentitySvc{
			resp: &identitypb.GetCurrentUserResponse{
				Data: &commonpb.UserInfo{Id: "100", Name: "u"},
			},
		},
		bundle: nil,
	}
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.Background())
	scopeInfo, result := u.GetScopeInfo(req)
	if result.Code != domain.AuthSuccess {
		t.Fatalf("GetScopeInfo expected AuthSuccess, got code %d detail %q", result.Code, result.Detail)
	}
	if scopeInfo.OrgID != 0 {
		t.Errorf("GetScopeInfo with no org header expected OrgID 0, got %d", scopeInfo.OrgID)
	}
}

func TestUserState_GetInfo_ReturnsCommonUserInfo(t *testing.T) {
	u := &userState{
		state:     GetInit,
		credStore: &mockCredStore{cred: &PersistedCredential{AccessToken: "at"}},
		IdentitySvc: &mockIdentitySvc{
			resp: &identitypb.GetCurrentUserResponse{
				Data: &commonpb.UserInfo{Id: "1", Name: "a", Nick: "b"},
			},
		},
		bundle: nil,
	}
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.Background())
	info, _ := u.GetInfo(req)
	// Ensure returned type is *common.UserInfo (wrapper around proto)
	var _ *common.UserInfo = info
	if info.UserInfo == nil || info.UserInfo.Id != "1" {
		t.Errorf("expected common.UserInfo wrapping Id=1, got %+v", info)
	}
}
