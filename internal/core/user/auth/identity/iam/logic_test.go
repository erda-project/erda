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

package iam

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	identitypb "github.com/erda-project/erda-proto-go/core/user/identity/pb"
	"github.com/erda-project/erda/internal/core/user/impl/iam"
)

// grant token: GET /iam/api/v1/admin/user/me with Bearer token
// response: { "success": true, "data": { "id": 1, "username": "u", ... } }
func TestGetCurrentUser_Grant(t *testing.T) {
	respBody := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"id":       float64(100),
			"username": "testuser",
			"nickname": "Test",
			"mobile":   "13800138000",
			"email":    "test@example.com",
			"avatar":   "https://avatar",
		},
	}
	bodyBytes, _ := json.Marshal(respBody)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/iam/api/v1/admin/user/me" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer token-abc" {
			t.Errorf("unexpected Authorization %s", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}))
	defer srv.Close()

	p := &provider{Cfg: &Config{BackendHost: srv.URL}}
	ctx := context.Background()
	req := &identitypb.GetCurrentUserRequest{
		Source:      identitypb.TokenSource_Grant,
		AccessToken: "token-abc",
	}
	resp, err := p.GetCurrentUser(ctx, req)
	if err != nil {
		t.Fatalf("GetCurrentUser: %v", err)
	}
	if resp.Data == nil {
		t.Fatal("expected non-nil Data")
	}
	if resp.Data.Id != "100" {
		t.Errorf("expected Id 100, got %s", resp.Data.Id)
	}
	if resp.Data.Name != "testuser" || resp.Data.Nick != "Test" {
		t.Errorf("expected Name=testuser Nick=Test, got Name=%s Nick=%s", resp.Data.Name, resp.Data.Nick)
	}
	if resp.Data.Phone != "13800138000" || resp.Data.Email != "test@example.com" {
		t.Errorf("expected Phone/Email, got Phone=%s Email=%s", resp.Data.Phone, resp.Data.Email)
	}
	if resp.SessionRefresh != nil {
		t.Error("Grant source should not return SessionRefresh")
	}
}

func TestGetCurrentUser_Grant_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer srv.Close()

	p := &provider{Cfg: &Config{BackendHost: srv.URL}}
	ctx := context.Background()
	req := &identitypb.GetCurrentUserRequest{
		Source:      identitypb.TokenSource_Grant,
		AccessToken: "bad",
	}
	_, err := p.GetCurrentUser(ctx, req)
	if err == nil {
		t.Fatal("expected error on HTTP 401")
	}
}

// cookie: GET /{app}/iam/api/v1/admin/user/find-by-token?token=value
func TestGetCurrentUser_Cookie(t *testing.T) {
	respBody := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"user": map[string]interface{}{
				"id":       float64(200),
				"username": "cookieuser",
				"nickname": "Cookie",
				"mobile":   "13900139000",
				"email":    "cookie@example.com",
				"avatar":   "/avatar.png",
			},
			"newToken": "",
			"expire":   float64(0),
		},
	}
	bodyBytes, _ := json.Marshal(respBody)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/erda/iam/api/v1/admin/user/find-by-token" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if r.URL.Query().Get("token") != "cookie-value" {
			t.Errorf("unexpected token query %s", r.URL.Query().Get("token"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}))
	defer srv.Close()

	cookieName := "sid"
	p := &provider{Cfg: &Config{BackendHost: srv.URL, ApplicationName: "erda"}}
	ctx := context.Background()
	req := &identitypb.GetCurrentUserRequest{
		Source:      identitypb.TokenSource_Cookie,
		AccessToken: "cookie-value",
		CookieName:  &cookieName,
	}
	resp, err := p.GetCurrentUser(ctx, req)
	if err != nil {
		t.Fatalf("GetCurrentUser: %v", err)
	}
	if resp.Data == nil {
		t.Fatal("expected non-nil Data")
	}
	if resp.Data.Id != "200" || resp.Data.Name != "cookieuser" || resp.Data.Nick != "Cookie" {
		t.Errorf("unexpected user: %+v", resp.Data)
	}
	if resp.SessionRefresh != nil {
		t.Error("no newToken in response, SessionRefresh should be nil")
	}
}

func TestGetCurrentUser_Cookie_WithRefresh(t *testing.T) {
	respBody := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"user": map[string]interface{}{
				"id":       float64(1),
				"username": "u",
				"nickname": "n",
				"mobile":   "",
				"email":    "",
			},
			"newToken": "new-session-token",
			"expire":   float64(3600),
			"cookieConfig": map[string]interface{}{
				"domain":   ".example.com",
				"path":     "/",
				"secure":   true,
				"httpOnly": true,
			},
		},
	}
	bodyBytes, _ := json.Marshal(respBody)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}))
	defer srv.Close()

	cookieName := "sid"
	p := &provider{Cfg: &Config{BackendHost: srv.URL, ApplicationName: "erda"}}
	ctx := context.Background()
	req := &identitypb.GetCurrentUserRequest{
		Source:      identitypb.TokenSource_Cookie,
		AccessToken: "old-token",
		CookieName:  &cookieName,
	}
	resp, err := p.GetCurrentUser(ctx, req)
	if err != nil {
		t.Fatalf("GetCurrentUser: %v", err)
	}
	if resp.SessionRefresh == nil || resp.SessionRefresh.Cookie == nil {
		t.Fatal("expected SessionRefresh with Cookie when newToken present")
	}
	c := resp.SessionRefresh.Cookie
	if c.Name != "sid" || c.Value != "new-session-token" {
		t.Errorf("expected cookie name=sid value=new-session-token, got name=%s value=%s", c.Name, c.Value)
	}
	if c.Path != "/" {
		t.Errorf("expected path /, got %s", c.Path)
	}
	if c.Domain != ".example.com" {
		t.Errorf("expected domain .example.com, got %s", c.Domain)
	}
	if c.HttpOnly == nil || !*c.HttpOnly {
		t.Error("expected HttpOnly true")
	}
	if c.Secure == nil || !*c.Secure {
		t.Error("expected Secure true")
	}
	if c.ExpireAt == nil {
		t.Error("expected ExpireAt set when expire > 0")
	}
	if c.MaxAge != 3600 {
		t.Errorf("expected MaxAge 3600, got %d", c.MaxAge)
	}
}

func TestGetCurrentUser_Cookie_NilCookieName(t *testing.T) {
	p := &provider{Cfg: &Config{BackendHost: "http://iam", ApplicationName: "erda"}}
	ctx := context.Background()
	req := &identitypb.GetCurrentUserRequest{
		Source:      identitypb.TokenSource_Cookie,
		AccessToken: "x",
		CookieName:  nil,
	}
	_, err := p.GetCurrentUser(ctx, req)
	if err == nil {
		t.Fatal("expected error when CookieName is nil")
	}
	if err.Error() != "illegal cookie name" {
		t.Errorf("expected 'illegal cookie name', got %q", err.Error())
	}
}

func TestGetCurrentUser_Cookie_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("forbidden"))
	}))
	defer srv.Close()

	cookieName := "sid"
	p := &provider{Cfg: &Config{BackendHost: srv.URL, ApplicationName: "erda"}}
	ctx := context.Background()
	req := &identitypb.GetCurrentUserRequest{
		Source:      identitypb.TokenSource_Cookie,
		AccessToken: "x",
		CookieName:  &cookieName,
	}
	_, err := p.GetCurrentUser(ctx, req)
	if err == nil {
		t.Fatal("expected error on HTTP 403")
	}
}

// Ensure impl/iam response shape is used (logic decodes into iam.Response[iam.UserDto] etc.)
func Test_getUserWithGrantedToken_mapsFields(t *testing.T) {
	dto := iam.UserDto{
		ID:       999,
		Username: "u",
		Nickname: "n",
		Mobile:   "m",
		Email:    "e",
		Avatar:   "a",
	}
	wrapper := iam.Response[iam.UserDto]{Success: true, Data: dto}
	bodyBytes, _ := json.Marshal(wrapper)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}))
	defer srv.Close()

	p := &provider{Cfg: &Config{BackendHost: srv.URL}}
	user, err := p.getUserWithGrantedToken("t")
	if err != nil {
		t.Fatalf("getUserWithGrantedToken: %v", err)
	}
	if user.Id != "999" {
		t.Errorf("expected Id 999, got %s", user.Id)
	}
	// grant path maps Username->Name, Nickname->Nick, Mobile->Phone, Email; Avatar not set in logic
	if user.Name != "u" || user.Nick != "n" || user.Phone != "m" || user.Email != "e" {
		t.Errorf("unexpected mapping: %+v", user)
	}
}

func Test_getUserWithCookie_mapsUserAndRefresh(t *testing.T) {
	wt := iam.UserWithToken{
		User: iam.UserDto{
			ID:       88,
			Username: "cu",
			Nickname: "CN",
			Mobile:   "100",
			Email:    "c@e.com",
			Avatar:   "av",
		},
		NewToken: "nt",
		Expire:   7200,
		CookieConfig: &iam.CookieConfig{
			Domain:   "d",
			Path:     "/p",
			Secure:   true,
			HttpOnly: true,
		},
	}
	wrapper := iam.Response[iam.UserWithToken]{Success: true, Data: wt}
	bodyBytes, _ := json.Marshal(wrapper)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}))
	defer srv.Close()

	name := "sid"
	p := &provider{Cfg: &Config{BackendHost: srv.URL, ApplicationName: "erda"}}
	user, refresh, err := p.getUserWithCookie(&name, "v")
	if err != nil {
		t.Fatalf("getUserWithCookie: %v", err)
	}
	if user.Id != "88" || user.Name != "cu" || user.Nick != "CN" || user.Avatar != "av" {
		t.Errorf("unexpected user: %+v", user)
	}
	if refresh == nil || refresh.Cookie == nil {
		t.Fatal("expected refresh cookie")
	}
	if refresh.Cookie.Name != "sid" || refresh.Cookie.Value != "nt" {
		t.Errorf("unexpected cookie: %+v", refresh.Cookie)
	}
	if refresh.Cookie.Domain != "d" || refresh.Cookie.Path != "/p" {
		t.Errorf("unexpected domain/path: %s %s", refresh.Cookie.Domain, refresh.Cookie.Path)
	}
}
