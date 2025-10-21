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

package akutil

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	clienttokenpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_token/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers"
)

func (util *AKUtil) TokenToClient(token string) (*clienttokenpb.ClientToken, *clientpb.Client, error) {
	_, allTokensV, err := util.cache.ListAll(context.Background(), cachetypes.ItemTypeClientToken)
	if err != nil {
		return nil, nil, err
	}
	var matchedTokens []*clienttokenpb.ClientToken
	for _, clientToken := range allTokensV.([]*clienttokenpb.ClientToken) {
		if clientToken.Token == token {
			matchedTokens = append(matchedTokens, clientToken)
		}
	}
	if len(matchedTokens) == 0 {
		return nil, nil, fmt.Errorf("invalid token")
	}
	if len(matchedTokens) > 1 {
		return nil, nil, fmt.Errorf("multiple tokens found")
	}
	clientToken := matchedTokens[0]
	if isTokenExpired(clientToken) {
		return nil, nil, handlers.ErrTokenExpired
	}
	clientV, err := util.cache.GetByID(context.Background(), cachetypes.ItemTypeClient, clientToken.ClientId)
	if err != nil {
		return nil, nil, err
	}
	return clientToken, clientV.(*clientpb.Client), err
}

const EnvKeyEnableTokenExpireCheck = "ENABLE_TOKEN_EXPIRE_CHECK"

func isTokenExpired(token *clienttokenpb.ClientToken) bool {
	// only enable expire check when env is set to true
	v := os.Getenv(EnvKeyEnableTokenExpireCheck)
	if v == "" || strings.ToLower(v) != "true" {
		return false
	}
	if token == nil {
		return true
	}
	if token.ExpireAt == nil {
		return false
	}
	expireAt := token.ExpireAt.AsTime()
	// consider tokens with an expiration that is not in the future as expired
	return expireAt.Before(time.Now())
}
