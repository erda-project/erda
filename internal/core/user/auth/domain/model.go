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

import "net/http"

const (
	OAuthProviderUC  = "uc"
	OAuthProviderIAM = "iam"
)

const (
	Unauthed        = http.StatusUnauthorized
	AuthFail        = http.StatusForbidden
	InternalAuthErr = http.StatusInternalServerError
	AuthSuccess     = http.StatusOK
)

type OAuthToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	// Optional
	RefreshToken string `json:"refresh_token,omitempty"`
}

type AuthCredential struct {
	OAuthToken *OAuthToken
	JWTToken   string
	SessionID  string
}

type PersistedCredential struct {
	Authenticator RequestAuthenticator
	AccessToken   string
	// Optional
	SessionID string
}

type UserAuthResult struct {
	Code   int
	Detail string
}
