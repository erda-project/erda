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

package reverseproxy

import (
	"context"
	"time"
)

var (
	_ context.Context = (*Context)(nil)
)

type (
	ProvidersCtxKey    struct{ ProvidersCtxKey any }
	FiltersCtxKey      struct{ FiltersCtxKey any }
	ProviderCtxKey     struct{ ProviderCtxKey any }
	DBCtxKey           struct{ DBCtxKey any }
	LoggerCtxKey       struct{ LoggerCtxKey any }
	ReplacedPathCtxKey struct{ ReplacedPathCtxKey any }
	AddQueriesCtxKey   struct{ AddQueriesCtxKey any }
	LogHttpCtxKey      struct{ LogHttpCtxKey any }
	MutexCtxKey        struct{ MutexCtxKey any }
)

type Context struct {
	inner context.Context
}

func NewContext(values map[any]any) *Context {
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

func (c *Context) Clone() *Context {
	return &Context{inner: c.inner}
}

func WithValue(ctx *Context, key any, value any) {
	ctx.inner = context.WithValue(ctx.inner, key, value)
}
