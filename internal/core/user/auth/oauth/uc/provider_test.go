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

package uc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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
}

func TestProvider_convertExpiresIn2Time(t *testing.T) {
	p := &provider{Config: &Config{TokenCacheEarlyExpireRate: 0.8}}
	d := p.convertExpiresIn2Time(100)
	expected := time.Duration(80) * time.Second
	if d != expected {
		t.Errorf("expected %v, got %v", expected, d)
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
	// redirect_uri in UC is "base?referer=..."
	if resp.Data != "https://example.com/oauth/authorize?client_id=client1&redirect_uri=https%3A%2F%2Fapp%2Fcallback%3Freferer%3D%252Fdashboard&response_type=code&scope=public_profile" {
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
}

func TestProvider_ExchangeCode(t *testing.T) {
	body := `{"success":true,"access_token":"at","token_type":"Bearer","expires_in":3600,"refresh_token":"rt"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/token" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
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
	body := `{"success":true,"access_token":"at","token_type":"Bearer","expires_in":3600}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	p := &provider{
		Log: logrusx.New(),
		Config: &Config{
			BackendHost:  srv.URL,
			ClientID:     "c",
			ClientSecret: "s",
		},
	}
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
	body := `{"success":true,"access_token":"at","token_type":"Bearer","expires_in":3600}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
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
	p.tokenCache = gcache.New(10).LRU().Build()
	ctx := context.Background()
	resp, err := p.ExchangeClientCredentials(ctx, &pb.ExchangeClientCredentialsRequest{Refresh: true})
	if err != nil {
		t.Fatal(err)
	}
	if resp.AccessToken != "at" {
		t.Errorf("unexpected token: %+v", resp)
	}
}

func TestProvider_doExchange_UCError(t *testing.T) {
	body := `{"success":false,"error":"invalid_grant"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	p := &provider{
		Log: logrusx.New(),
		Config: &Config{
			BackendHost:  srv.URL,
			ClientID:     "c",
			ClientSecret: "s",
		},
	}
	_, err := p.doExchange(url.Values{})
	if err == nil {
		t.Fatal("expected error for success:false")
	}
}

func TestDecodeUCFlat(t *testing.T) {
	type T struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	body := `{"success":true,"access_token":"at","expires_in":3600}`
	result, err := DecodeUCFlat[T]([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	if result.AccessToken != "at" || result.ExpiresIn != 3600 {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestDecodeUCFlat_ErrorResponse(t *testing.T) {
	type T struct {
		A string `json:"a"`
	}
	body := `{"success":false,"error":"bad request"}`
	_, err := DecodeUCFlat[T]([]byte(body))
	if err == nil {
		t.Fatal("expected error for success:false")
	}
	if err.Error() != "bad request" {
		t.Errorf("expected error message 'bad request', got %q", err.Error())
	}
}

func TestDecodeUCFlat_InvalidJSON(t *testing.T) {
	type T struct {
		A string `json:"a"`
	}
	_, err := DecodeUCFlat[T]([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestDecodeUCFlat_SuccessNil(t *testing.T) {
	// when success is nil, we don't treat as error (UC may omit it)
	type T struct {
		V int `json:"v"`
	}
	body := `{"v":1}`
	result, err := DecodeUCFlat[T]([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	if result.V != 1 {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestDecodeUCFlat_SuccessTrueExplicit(t *testing.T) {
	type T struct {
		V int `json:"v"`
	}
	success := true
	body, _ := json.Marshal(struct {
		Success *bool `json:"success"`
		V       int   `json:"v"`
	}{&success, 42})
	result, err := DecodeUCFlat[T](body)
	if err != nil {
		t.Fatal(err)
	}
	if result.V != 42 {
		t.Errorf("unexpected result: %+v", result)
	}
}
