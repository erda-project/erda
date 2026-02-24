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

package sessionrefresh

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	identitypb "github.com/erda-project/erda-proto-go/core/user/identity/pb"
	"github.com/erda-project/erda/pkg/pointer"
)

type CookieDefaults struct {
	Name     string
	Domains  []string
	Path     string
	HTTPOnly bool
	SameSite http.SameSite
	MaxAge   time.Duration
}

func BuildCookies(refresh *identitypb.SessionRefresh, defaults CookieDefaults, req *http.Request) ([]*http.Cookie, error) {
	if refresh == nil || refresh.Cookie == nil {
		return nil, nil
	}
	cookie := refresh.Cookie
	if cookie.Value == "" {
		return nil, nil
	}

	name := cookie.Name
	if name == "" {
		name = defaults.Name
	}
	if name == "" {
		return nil, fmt.Errorf("session refresh cookie name is empty")
	}

	path := cookie.Path
	if path == "" {
		path = defaults.Path
	}
	if path == "" {
		path = "/"
	}

	domains := normalizeDomains(cookie.Domain, defaults.Domains)
	if len(domains) == 0 {
		domains = []string{""}
	}

	httpOnly := pointer.BoolDeref(cookie.HttpOnly, defaults.HTTPOnly)
	secureDefault := req != nil && req.TLS != nil
	secure := pointer.BoolDeref(cookie.Secure, secureDefault)

	sameSite := SameSiteToHTTP(cookie.SameSite)
	if sameSite == 0 {
		if defaults.SameSite > 0 {
			sameSite = defaults.SameSite
		} else {
			sameSite = http.SameSiteLaxMode
		}
	}

	maxAge := int(cookie.MaxAge)
	if maxAge == 0 && defaults.MaxAge > 0 {
		maxAge = int(defaults.MaxAge.Seconds())
	}

	var expires time.Time
	if cookie.ExpireAt != nil {
		expires = cookie.ExpireAt.AsTime()
	} else if defaults.MaxAge > 0 {
		expires = time.Now().Add(defaults.MaxAge)
	}

	var cookies []*http.Cookie
	for _, domain := range domains {
		c := &http.Cookie{
			Name:     name,
			Value:    cookie.Value,
			Path:     path,
			Domain:   domain,
			HttpOnly: httpOnly,
			Secure:   secure,
			SameSite: sameSite,
		}
		if !expires.IsZero() {
			c.Expires = expires
		}
		if maxAge != 0 {
			c.MaxAge = maxAge
		}
		cookies = append(cookies, c)
	}
	return cookies, nil
}

func normalizeDomains(domain string, defaults []string) []string {
	if strings.TrimSpace(domain) != "" {
		return []string{strings.TrimSpace(domain)}
	}
	if len(defaults) == 0 {
		return nil
	}
	out := make([]string, 0, len(defaults))
	for _, d := range defaults {
		d = strings.TrimSpace(d)
		if d != "" {
			out = append(out, d)
		}
	}
	return out
}

func SameSiteFromHTTP(s http.SameSite) identitypb.CookieSameSite {
	switch s {
	case http.SameSiteStrictMode:
		return identitypb.CookieSameSite_Strict
	case http.SameSiteNoneMode:
		return identitypb.CookieSameSite_None
	case http.SameSiteLaxMode:
		return identitypb.CookieSameSite_Lax
	default:
		return identitypb.CookieSameSite_Unspecified
	}
}

func SameSiteToHTTP(s identitypb.CookieSameSite) http.SameSite {
	switch s {
	case identitypb.CookieSameSite_Strict:
		return http.SameSiteStrictMode
	case identitypb.CookieSameSite_None:
		return http.SameSiteNoneMode
	case identitypb.CookieSameSite_Lax:
		return http.SameSiteLaxMode
	default:
		return 0
	}
}
