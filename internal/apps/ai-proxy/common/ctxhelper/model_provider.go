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

	modelproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

func GetModelProvider(ctx context.Context) (*modelproviderpb.ModelProvider, bool) {
	value, ok := ctx.Value(reverseproxy.CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyModelProvider{})
	if !ok || value == nil {
		return nil, false
	}
	prov, ok := value.(*modelproviderpb.ModelProvider)
	if !ok {
		return nil, false
	}
	return prov, true
}

func MustGetModelProvider(ctx context.Context) *modelproviderpb.ModelProvider {
	prov, ok := GetModelProvider(ctx)
	if !ok {
		panic("model provider not found in context")
	}
	return prov
}
