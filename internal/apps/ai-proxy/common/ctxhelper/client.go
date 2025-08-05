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

package ctxhelper

import (
	"context"
	"sync"

	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

func GetClient(ctx context.Context) (*clientpb.Client, bool) {
	value, ok := ctx.Value(CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyClient{})
	if !ok || value == nil {
		return nil, false
	}
	client, ok := value.(*clientpb.Client)
	if !ok {
		return nil, false
	}
	return client, true
}
