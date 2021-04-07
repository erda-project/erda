// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package endpoints

import (
	"context"
	"net/http"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

type endpoint func(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error)

func auth(f endpoint) endpoint {
	return func(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
		// TODO: auth
		return f(ctx, r, vars)
	}
}

func i18nPrinter(f endpoint) endpoint {
	return func(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
		lang := r.Header.Get("Lang")
		p := message.NewPrinter(language.English)
		if strutil.Contains(lang, "zh") {
			p = message.NewPrinter(language.SimplifiedChinese)
		}
		ctx2 := context.WithValue(ctx, "i18nPrinter", p)
		return f(ctx2, r, vars)
	}
}
