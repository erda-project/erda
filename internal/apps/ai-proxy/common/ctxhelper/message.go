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

	"github.com/erda-project/erda/internal/apps/ai-proxy/models/message"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

func GetMessageGroup(ctx context.Context) (*message.Group, bool) {
	value, ok := ctx.Value(CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyMessageGroup{})
	if !ok || value == nil {
		return nil, false
	}
	mg, ok := value.(*message.Group)
	if !ok {
		return nil, false
	}
	return mg, true
}

func PutMessageGroup(ctx context.Context, mg message.Group) {
	m := ctx.Value(CtxKeyMap{}).(*sync.Map)
	m.Store(vars.MapKeyMessageGroup{}, &mg)
}
