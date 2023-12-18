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
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

func IsAdmin(ctx context.Context) bool {
	isAdmin, ok := ctx.Value(vars.CtxKeyIsAdmin{}).(bool)
	if !ok {
		return false
	}
	return isAdmin
}

func IsClient(ctx context.Context) bool {
	clientId, ok := ctx.Value(vars.CtxKeyClientId{}).(string)
	if !ok {
		return false
	}
	return clientId != ""
}

func GetClientId(ctx context.Context) string {
	return ctx.Value(vars.CtxKeyClientId{}).(string)
}

func GetClient(ctx context.Context) *clientpb.Client {
	return ctx.Value(vars.CtxKeyClient{}).(*clientpb.Client)
}

func Valid(ctx context.Context) bool {
	if IsAdmin(ctx) {
		return true
	}
	return IsClient(ctx)
}
