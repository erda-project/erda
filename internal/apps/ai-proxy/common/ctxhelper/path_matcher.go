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

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/router_define/path_matcher"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

func MustGetPathMatcher(ctx context.Context) *path_matcher.PathMatcher {
	pm, ok := GetPathMatcher(ctx)
	if !ok {
		panic("path matcher not found in context")
	}
	return pm
}

func GetPathMatcher(ctx context.Context) (*path_matcher.PathMatcher, bool) {
	v := ctx.Value(vars.CtxKeyPathMatcher{})
	if v == nil {
		return nil, false
	}
	pm, ok := v.(*path_matcher.PathMatcher)
	return pm, ok
}

func GetPathParam(ctx context.Context, key string) (string, bool) {
	pm, ok := GetPathMatcher(ctx)
	if !ok {
		return "", false
	}
	return pm.GetValue(key)
}
