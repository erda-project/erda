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

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/logs/logrusx"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

func MustGetLogger(ctx context.Context) logs.Logger {
	logger, ok := GetLogger(ctx)
	if ok {
		return logger
	}
	logger = logrusx.New()
	PutLogger(ctx, logger)
	return logger
}

func GetLogger(ctx context.Context) (logs.Logger, bool) {
	value, ok := ctx.Value(CtxKeyMap{}).(*sync.Map).Load(vars.MapKeyLogger{})
	if !ok || value == nil {
		return nil, false
	}
	logger, ok := value.(logs.Logger)
	if !ok {
		return nil, false
	}
	return logger, true
}

func PutLogger(ctx context.Context, logger logs.Logger) {
	m := ctx.Value(CtxKeyMap{}).(*sync.Map)
	m.Store(vars.MapKeyLogger{}, logger)
}
