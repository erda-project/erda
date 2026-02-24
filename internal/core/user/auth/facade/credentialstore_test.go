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

	identitypb "github.com/erda-project/erda-proto-go/core/user/identity/pb"
)

func TestNewCookieStore(t *testing.T) {
	store := NewCookieStore("sid")
	if store == nil {
		t.Fatal("NewCookieStore should not return nil")
	}
}

func TestCookieStore_Load_withCookie(t *testing.T) {
	store := NewCookieStore("sid")
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "sid", Value: "token-abc"})

	ctx := context.Background()
	cred, err := store.Load(ctx, req)
	if err != nil {
		t.Fatalf("Load with cookie: %v", err)
	}
	if cred == nil {
		t.Fatal("expected non-nil credential")
	}
	if cred.Type != identitypb.TokenSource_Cookie {
		t.Errorf("expected Type Cookie, got %v", cred.Type)
	}
	if cred.AccessToken != "token-abc" {
		t.Errorf("expected AccessToken token-abc, got %q", cred.AccessToken)
	}
	if cred.CookieName != "sid" {
		t.Errorf("expected CookieName sid, got %q", cred.CookieName)
	}
}

func TestCookieStore_Load_noCookie(t *testing.T) {
	store := NewCookieStore("sid")
	req, _ := http.NewRequest(http.MethodGet, "/", nil)

	ctx := context.Background()
	cred, err := store.Load(ctx, req)
	if err == nil {
		t.Fatal("expected error when cookie is missing")
	}
	if cred != nil {
		t.Errorf("expected nil credential on error, got %+v", cred)
	}
}

func TestCookieStore_Load_wrongCookieName(t *testing.T) {
	store := NewCookieStore("sid")
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "other", Value: "x"})

	ctx := context.Background()
	_, err := store.Load(ctx, req)
	if err == nil {
		t.Fatal("expected error when requested cookie name is missing")
	}
}

func TestToGetCurrentUserRequest_withCookieName(t *testing.T) {
	cred := &PersistedCredential{
		Type:        identitypb.TokenSource_Cookie,
		AccessToken: "at",
		CookieName:  "sid",
	}
	req := ToGetCurrentUserRequest(cred)
	if req == nil {
		t.Fatal("expected non-nil request")
	}
	if req.AccessToken != "at" {
		t.Errorf("expected AccessToken at, got %q", req.AccessToken)
	}
	if req.Source != identitypb.TokenSource_Cookie {
		t.Errorf("expected Source Cookie, got %v", req.Source)
	}
	if req.CookieName == nil || *req.CookieName != "sid" {
		t.Errorf("expected CookieName sid, got %v", req.CookieName)
	}
}

func TestToGetCurrentUserRequest_withoutCookieName(t *testing.T) {
	cred := &PersistedCredential{
		Type:        identitypb.TokenSource_Grant,
		AccessToken: "at",
		CookieName:  "",
	}
	req := ToGetCurrentUserRequest(cred)
	if req == nil {
		t.Fatal("expected non-nil request")
	}
	if req.AccessToken != "at" {
		t.Errorf("expected AccessToken at, got %q", req.AccessToken)
	}
	if req.Source != identitypb.TokenSource_Grant {
		t.Errorf("expected Source Grant, got %v", req.Source)
	}
	if req.CookieName != nil {
		t.Errorf("expected nil CookieName when empty, got %v", req.CookieName)
	}
}
