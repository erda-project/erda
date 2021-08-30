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

package token

import (
	"net/http"

	openapiauth "github.com/erda-project/erda/modules/core/openapi-ng/auth"
	"github.com/erda-project/erda/modules/openapi/api/spec"
	"github.com/erda-project/erda/modules/openapi/auth"
)

func (p *provider) Weight() int64 { return p.Cfg.Weight }

func (p *provider) Match(r *http.Request, opts openapiauth.Options) (bool, interface{}) {
	check, _ := opts.Get("CheckToken").(bool)
	if check {
		if authorization := r.Header.Get("Authorization"); len(authorization) > 0 {
			return true, authorization
		}
	}
	return false, nil
}

func (p *provider) Check(r *http.Request, data interface{}, opts openapiauth.Options) (bool, *http.Request, error) {
	authorization := data.(string)
	client, err := p.checkToken(nil, r, authorization)
	if err != nil {
		return false, r, nil
	}
	r.Header.Set("Client-ID", client.ClientID)
	r.Header.Set("Client-Name", client.ClientName)
	return true, r, nil
}

type clientToken struct {
	ClientID   string
	ClientName string
}

// checkToken try:
// 1. uc token
// 2. openapi oauth2 token
func (p *provider) checkToken(spec *spec.Spec, req *http.Request, token string) (clientToken, error) {
	// 1. uc token
	ucToken, err := auth.VerifyUCClientToken(token)
	if err == nil {
		return clientToken{
			ClientID:   ucToken.ClientID,
			ClientName: ucToken.ClientName,
		}, nil
	}
	// 2. openapi oauth2 token
	oauthToken, err := auth.VerifyOpenapiOAuth2Token(p.oauth2server, nil, req)
	if err != nil {
		return clientToken{}, err
	}
	return clientToken{
		ClientID:   oauthToken.ClientID,
		ClientName: oauthToken.ClientName,
	}, nil
}
