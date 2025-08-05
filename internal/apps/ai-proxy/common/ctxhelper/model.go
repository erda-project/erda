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

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

func GetModel(ctx context.Context) (*modelpb.Model, bool) {
	value, ok := ctx.Value(CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyModel{})
	if !ok || value == nil {
		return nil, false
	}
	model, ok := value.(*modelpb.Model)
	if !ok {
		return nil, false
	}
	return model, true
}

func MustGetModel(ctx context.Context) *modelpb.Model {
	model, ok := GetModel(ctx)
	if !ok {
		panic("model not found in context")
	}
	return model
}
