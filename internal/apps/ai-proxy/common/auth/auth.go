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

package auth

import (
	"context"

	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	clienttokenpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_token/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

func IsAdmin(ctx context.Context) bool {
	isAdmin, ok := ctxhelper.GetIsAdmin(ctx)
	if !ok {
		return false
	}
	return isAdmin
}

func IsClient(ctx context.Context) bool {
	client, _ := ctxhelper.GetClient(ctx)
	clientToken, _ := ctxhelper.GetClientToken(ctx)
	return client != nil && clientToken == nil
}

func IsClientToken(ctx context.Context) bool {
	clientToken, _ := ctxhelper.GetClientToken(ctx)
	return clientToken != nil
}

func GetClient(ctx context.Context) *clientpb.Client {
	client, _ := ctxhelper.GetClient(ctx)
	return client
}

func GetClientToken(ctx context.Context) *clienttokenpb.ClientToken {
	clientToken, _ := ctxhelper.GetClientToken(ctx)
	return clientToken
}

func IsLoggedIn(ctx context.Context) bool {
	return IsAdmin(ctx) || IsClient(ctx) || IsClientToken(ctx)
}

func IsAdminOrClient(ctx context.Context) bool {
	return IsAdmin(ctx) || IsClient(ctx)
}
