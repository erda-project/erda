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

	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	clienttokenpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_token/pb"
)

func (util *AKUtil) TokenToClient(token string) (*clienttokenpb.ClientToken, *clientpb.Client, error) {
	pagingResp, err := util.Dao.ClientTokenClient().Paging(context.Background(), &clienttokenpb.ClientTokenPagingRequest{
		PageSize: 1,
		PageNum:  1,
		Token:    token,
	})
	if err != nil {
		return nil, nil, err
	}
	if pagingResp.Total == 0 {
		return nil, nil, nil
	}
	if pagingResp.Total > 1 {
		return nil, nil, fmt.Errorf("multiple clients found by token: %s", token)
	}
	clientToken := pagingResp.List[0]
	client, err := util.Dao.ClientClient().Get(context.Background(), &clientpb.ClientGetRequest{
		ClientId: clientToken.ClientId,
	})
	return clientToken, client, err
}
