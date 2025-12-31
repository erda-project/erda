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

package util

import (
	"net/url"

	"github.com/erda-project/erda-proto-go/core/user/oauth/pb"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
)

func ConvertURLValuesToPb(params url.Values) map[string]*pb.StringList {
	result := make(map[string]*pb.StringList)
	for k, v := range params {
		sl := &pb.StringList{
			Values: make([]string, 0, len(v)),
		}
		for _, vv := range v {
			sl.Values = append(sl.Values, vv)
		}
		result[k] = sl
	}
	return result
}

func ConvertOAuthDomainToPb(oauthToken *domain.OAuthToken) *pb.OAuthToken {
	return &pb.OAuthToken{
		AccessToken:  oauthToken.AccessToken,
		RefreshToken: oauthToken.RefreshToken,
		ExpiresIn:    oauthToken.ExpiresIn,
		TokenType:    oauthToken.TokenType,
	}
}

func ConvertPbToOAuthDomain(oauthTokenPb *pb.OAuthToken) *domain.OAuthToken {
	return &domain.OAuthToken{
		AccessToken:  oauthTokenPb.AccessToken,
		TokenType:    oauthTokenPb.RefreshToken,
		ExpiresIn:    oauthTokenPb.ExpiresIn,
		RefreshToken: oauthTokenPb.TokenType,
	}
}
