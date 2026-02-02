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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bluele/gcache"
	"github.com/erda-project/erda-infra/base/logs/logrusx"
	"github.com/erda-project/erda-proto-go/core/user/oauth/pb"
)

func TestProvider_Init(t *testing.T) {
	p := &provider{
		Register: nil,
		Log:      logrusx.New(),
		Config: &Config{
			TokenCacheEarlyExpireRate: 0,
			TokenCacheSize:            100,
		},
	}
	err := p.Init(nil)
	if err != nil {
		t.Fatal(err)
	}
	if p.Config.TokenCacheEarlyExpireRate != defaultEarlyExpireRate {
		t.Errorf("expected early expire rate %v, got %v", defaultEarlyExpireRate, p.Config.TokenCacheEarlyExpireRate)
	}
	if p.tokenCache == nil {
		t.Error("tokenCache should be initialized")
	}

	// illegal rate >= 1
	p2 := &provider{
		Log:    logrusx.New(),
		Config: &Config{TokenCacheEarlyExpireRate: 1.5, TokenCacheSize: 100},
	}
	_ = p2.Init(nil)
	if p2.Config.TokenCacheEarlyExpireRate != defaultEarlyExpireRate {
		t.Errorf("expected default rate for >= 1, got %v", p2.Config.TokenCacheEarlyExpireRate)
	}
}

func TestProvider_convertExpiresIn2Time(t *testing.T) {
	p := &provider{Config: &Config{TokenCacheEarlyExpireRate: 0.8}}
	d := p.convertExpiresIn2Time(100)
	expected := time.Duration(80) * time.Second
	if d != expected {
		t.Errorf("expected %v, got %v", expected, d)
	}
}

func TestProvider_userTokenCacheKey(t *testing.T) {
	p := &provider{Config: &Config{OAuthTokenCacheSecret: "secret"}}
	k1 := p.userTokenCacheKey("user1", "pass1")
	k2 := p.userTokenCacheKey("user1", "pass1")
	if k1 != k2 {
		t.Error("same input should produce same cache key")
	}
	if k1 == p.userTokenCacheKey("user1", "pass2") {
		t.Error("different password should produce different key")
	}
	if len(k1) == 0 || k1[:len(userTokenCachePrefix)] != userTokenCachePrefix {
		t.Errorf("key should have prefix %q, got %q", userTokenCachePrefix, k1)
	}
}

func TestProvider_AuthURL(t *testing.T) {
	p := &provider{
		Log: logrusx.New(),
		Config: &Config{
			FrontendURL: "https://example.com",
			ClientID:    "client1",
			RedirectURI: "https://app/callback",
		},
	}
	ctx := context.Background()
	resp, err := p.AuthURL(ctx, &pb.AuthURLRequest{Referer: "/dashboard"})
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil || resp.Data == "" {
		t.Fatal("expected non-empty AuthURL")
	}
	if resp.Data != "https://example.com/iam/oauth2/server/authorize?client_id=client1&redirect_uri=https%3A%2F%2Fapp%2Fcallback&response_type=code&scope=api&state=%2Fdashboard" {
		t.Logf("AuthURL: %s", resp.Data)
	}
}

func TestProvider_LogoutURL(t *testing.T) {
	p := &provider{
		Log: logrusx.New(),
		Config: &Config{
			FrontendURL: "https://example.com",
			ClientID:    "client1",
			RedirectURI: "https://app/callback",
		},
	}
	ctx := context.Background()
	resp, err := p.LogoutURL(ctx, &pb.LogoutURLRequest{Referer: "/"})
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil || resp.Data == "" {
		t.Fatal("expected non-empty LogoutURL")
	}
	if resp.Data != "https://example.com/logout?redirectUrl=https%3A%2F%2Fexample.com%2Fiam%2Foauth2%2Fserver%2Fauthorize%3Fclient_id%3Dclient1%26redirect_uri%3Dhttps%253A%252F%252Fapp%252Fcallback%26response_type%3Dcode%26scope%3Dapi%26state%3D%252F" {
		t.Logf("LogoutURL: %s", resp.Data)
	}
}

func TestProvider_ExchangeCode(t *testing.T) {
	tokenBody := map[string]interface{}{
		"access_token":  "at",
		"token_type":    "Bearer",
		"expires_in":    int64(3600),
		"refresh_token": "rt",
	}
	bodyBytes, _ := json.Marshal(tokenBody)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/iam/oauth2/server/token" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}))
	defer srv.Close()

	p := &provider{
		Log: logrusx.New(),
		Config: &Config{
			BackendHost:  srv.URL,
			ClientID:     "c",
			ClientSecret: "s",
			RedirectURI:  "https://app/cb",
		},
	}
	p.tokenCache = newTestCache()
	ctx := context.Background()
	resp, err := p.ExchangeCode(ctx, &pb.ExchangeCodeRequest{Code: "code1"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.AccessToken != "at" || resp.RefreshToken != "rt" || resp.ExpiresIn != 3600 {
		t.Errorf("unexpected token: %+v", resp)
	}
}

func TestProvider_ExchangePassword(t *testing.T) {
	tokenBody := map[string]interface{}{
		"access_token": "at",
		"token_type":   "Bearer",
		"expires_in":   int64(3600),
	}
	bodyBytes, _ := json.Marshal(tokenBody)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}))
	defer srv.Close()

	p := &provider{
		Log: logrusx.New(),
		Config: &Config{
			BackendHost:              srv.URL,
			ClientID:                 "c",
			ClientSecret:             "s",
			UserTokenCacheEnabled:    false,
			UserTokenCacheExpireTime: time.Minute,
		},
	}
	p.tokenCache = newTestCache()
	ctx := context.Background()
	resp, err := p.ExchangePassword(ctx, &pb.ExchangePasswordRequest{Username: "u", Password: "p"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.AccessToken != "at" {
		t.Errorf("unexpected token: %+v", resp)
	}
}

func TestProvider_ExchangeClientCredentials(t *testing.T) {
	tokenBody := map[string]interface{}{
		"access_token": "at",
		"token_type":   "Bearer",
		"expires_in":   int64(3600),
	}
	bodyBytes, _ := json.Marshal(tokenBody)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}))
	defer srv.Close()

	p := &provider{
		Log: logrusx.New(),
		Config: &Config{
			BackendHost:               srv.URL,
			ClientID:                  "c",
			ClientSecret:              "s",
			ServerTokenCacheEnabled:   false,
			TokenCacheEarlyExpireRate: 0.8,
		},
	}
	p.tokenCache = newTestCache()
	ctx := context.Background()
	resp, err := p.ExchangeClientCredentials(ctx, &pb.ExchangeClientCredentialsRequest{Refresh: true})
	if err != nil {
		t.Fatal(err)
	}
	if resp.AccessToken != "at" {
		t.Errorf("unexpected token: %+v", resp)
	}
}

func TestProvider_doExchange_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer srv.Close()

	log := logrusx.New()
	log.SetOutput(io.Discard) // suppress ERRO in test output
	p := &provider{
		Log: log,
		Config: &Config{
			BackendHost:  srv.URL,
			ClientID:     "c",
			ClientSecret: "s",
		},
	}
	ctx := context.Background()
	_, err := p.doExchange(ctx, nil)
	if err == nil {
		t.Fatal("expected error for 401")
	}
}

// newTestCache returns a minimal cache for tests (avoids importing gcache in test for builder).
func newTestCache() gcache.Cache {
	return gcache.New(10).LRU().Build()
}
