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

package filter

import (
	"context"
	"fmt"
	"time"
)

var (
	_ context.Context = (*Context)(nil)
)

type (
	// RouteCtxKey 用以从 context.Context 中获取 route.Route 以获取 route.Route 的更多配置信息
	RouteCtxKey        struct{ RouteCtxKey any }
	ProvidersCtxKey    struct{ ProvidersCtxKey any }
	FiltersCtxKey      struct{ FiltersCtxKey any }
	ProviderCtxKey     struct{ ProviderCtxKey any }
	OperationCtxKey    struct{ OperationCtxKey any }
	DBCtxKey           struct{ DBCtxKey any }
	LoggerCtxKey       struct{ LoggerCtxKey any }
	ReplacedPathCtxKey struct{ ReplacedPathCtxKey any }
	AddQueriesCtxKey   struct{ AddQueriesCtxKey any }
	LogHttpCtxKey      struct{ LogHttpCtxKey any }
)

type Context struct {
	inner context.Context
}

func NewContext(values map[any]any) context.Context {
	var ctx = &Context{inner: context.Background()}
	for k, v := range values {
		WithValue(ctx, k, v)
	}
	return ctx
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.inner.Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.inner.Done()
}

func (c *Context) Err() error {
	return c.inner.Err()
}

func (c *Context) Value(key any) any {
	return c.inner.Value(key)
}

func WithValue(ctx context.Context, key any, value any) {
	if c, ok := ctx.(*Context); ok {
		c.inner = context.WithValue(c.inner, key, value)
		return
	}
	panic(fmt.Sprintf("the ctx context.Context must be type %T, got %T", new(Context), ctx))
}
