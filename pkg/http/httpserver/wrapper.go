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

package httpserver

import (
	"context"
	"net/http"

	"github.com/erda-project/erda-infra/providers/legacy/httpendpoints/i18n"
)

func Wrap(h Handler, wrappers ...HandlerWrapper) Handler {
	n := len(wrappers)
	if n == 0 {
		return h
	}
	for i := n - 1; i >= 0; i-- {
		h = wrappers[i](h)
	}
	return h
}

type HandlerWrapper func(handler Handler) Handler

func WithI18nCodes(h Handler) Handler {
	return func(ctx context.Context, r *http.Request, vars map[string]string) (Responser, error) {
		ctx = context.WithValue(ctx, "Lang", i18n.Language(r))
		return h(ctx, r, vars)
	}
}
