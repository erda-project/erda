// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package apistructs

type OpenapiOAuth2Token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type OpenapiOAuth2TokenPayload struct {
	AccessTokenExpiredIn string            `json:"accessTokenExpiredIn"` // such as "300ms", "-1.5h" or "2h45m". "0" means it doesn't expire.
	AllowAccessAllAPIs   bool              `json:"allowAccessAllApIs,omitempty"`
	AccessibleAPIs       []AccessibleAPI   `json:"accessibleAPIs,omitempty"`
	Metadata             map[string]string `json:"metadata,omitempty"`
}

type AccessibleAPI struct {
	Path   string `json:"path"`
	Method string `json:"method"`
	Schema string `json:"schema"`
}

type OpenapiOAuth2TokenGetRequest struct {
	ClientID     string
	ClientSecret string
	Payload      OpenapiOAuth2TokenPayload `json:"payload"`
}

type OpenapiOAuth2TokenGetResponse OpenapiOAuth2Token

type OpenapiOAuth2TokenInvalidateRequest struct {
	AccessToken string `json:"access_token"`
}

type OpenapiOauth2TokenInvalidateResponse OpenapiOAuth2Token
