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

package prehandle

import (
	"context"
	"net/http"
	"time"

	"github.com/erda-project/erda/modules/openapi/conf"
)

func ReplaceOldCookie(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	if conf.OldCookieDomain() == "" {
		return
	}
	cs := req.Cookies()
	count := 0
	for _, c := range cs {
		if c.Name == conf.SessionCookieName() {
			count++
		}
	}
	if count > 1 {
		http.SetCookie(rw, &http.Cookie{
			Name:     conf.SessionCookieName(),
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
			Domain:   conf.OldCookieDomain(),
			HttpOnly: true,
		})
	}
}
