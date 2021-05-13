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

package csrf

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/random"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/openapi-ng/common"
	"github.com/erda-project/erda/modules/openapi-ng/interceptors"
)

type config struct {
	Order             int
	AllowEmptyReferer bool          `file:"allow_empty_referer" default:"false"`
	TokenGenerator    string        `file:"token_generator" default:"random"` // random、random:
	TokenLookup       string        `file:"token_lookup"`                     // "header:<name>"、"form:<name>"、 "query:<name>"
	CookieName        string        `file:"cookie_name" default:"csrf" desc:"name of the CSRF cookie. This cookie will store CSRF token. optional."`
	CookieDomain      string        `file:"cookie_domain" desc:"domain of the CSRF cookie. optional."`
	CookiePath        string        `file:"cookie_path" desc:"path of the CSRF cookie. optional."`
	CookieMaxAge      time.Duration `file:"cookie_max_age" default:"24h" desc:"max age of the CSRF cookie. optional."`
	CookieSecure      bool          `file:"cookie_secure" desc:"indicates if CSRF cookie is secure. optional."`
	CookieHTTPOnly    bool          `file:"cookie_http_only" desc:"indicates if CSRF cookie is HTTP only. optional."`
}

type (
	csrfTokenExtractor func(r *http.Request) (string, error)
	tokenGenerator     struct {
		gen   func(r *http.Request) string
		valid func(cookieToken, clientToken string, r *http.Request) bool
	}
	contextKey struct{}
)

// GetToken .
func GetToken(ctx context.Context) string {
	token, _ := ctx.Value(contextKey{}).(string)
	return token
}

// +provider
type provider struct {
	Cfg       *config
	generator *tokenGenerator
	extractor csrfTokenExtractor
}

func (p *provider) Init(ctx servicehub.Context) error {
	// TokenLookup
	if len(p.Cfg.TokenLookup) <= 0 {
		p.Cfg.TokenLookup = "header:X-CSRF-Token"
	}
	parts := strings.Split(p.Cfg.TokenLookup, ":")
	tokenName := parts[1]
	if len(tokenName) <= 0 {
		return errors.New("token name must not be empty")
	}
	switch parts[0] {
	case "form":
		p.extractor = csrfTokenFromForm(tokenName)
	case "query":
		p.extractor = csrfTokenFromQuery(tokenName)
	case "header":
		p.extractor = csrfTokenFromHeader(tokenName)
	default:
		return fmt.Errorf("invalid token_lookup %q", p.Cfg.TokenLookup)
	}

	// TokenGenerator
	parts = strings.Split(p.Cfg.TokenGenerator, ":")
	gen, ok := tokenGenerators[parts[0]]
	if !ok {
		return fmt.Errorf("invalid token_generator %q", p.Cfg.TokenGenerator)
	}
	var args []string
	if len(parts) > 1 && len(parts[1]) > 0 {
		args = strings.Split(parts[1], ",")
	}
	generator, err := gen(p.Cfg, args)
	if err != nil {
		return fmt.Errorf("fail to create token generator: %s", err)
	}
	p.generator = generator
	return nil
}

func (p *provider) List() []*interceptors.Interceptor {
	return []*interceptors.Interceptor{
		{Order: p.Cfg.Order, Wrapper: p.Interceptor},
	}
}

// Interceptor .
func (p *provider) Interceptor(h http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if p.Cfg.AllowEmptyReferer {
			referer := r.Header.Get("Referer")
			if referer == "" {
				h(rw, r)
				return
			}
		}
		k, err := r.Cookie(p.Cfg.CookieName)
		var token string
		if err == nil {
			// Reuse token
			token = k.Value
		}
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		default:
			// Validate token only for requests which are not defined as 'safe' by RFC7231
			clientToken, err := p.extractor(r)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				common.WriteError(err, rw)
				return
			}
			if !p.generator.valid(token, clientToken, r) {
				rw.WriteHeader(http.StatusForbidden)
				common.WriteError(errors.New("invalid csrf token"), rw)
				return
			}
		}
		if len(token) <= 0 {
			// Generate token
			token = p.generator.gen(r)
		}

		// Set CSRF cookie
		cookie := new(http.Cookie)
		cookie.Name = p.Cfg.CookieName
		cookie.Value = token
		if p.Cfg.CookiePath != "" {
			cookie.Path = p.Cfg.CookiePath
		}
		if p.Cfg.CookieDomain != "" {
			cookie.Domain = p.Cfg.CookieDomain
		}
		cookie.Expires = time.Now().Add(p.Cfg.CookieMaxAge)
		cookie.Secure = p.Cfg.CookieSecure
		cookie.HttpOnly = p.Cfg.CookieHTTPOnly
		http.SetCookie(rw, cookie)

		// Store token in the context
		r = r.WithContext(context.WithValue(r.Context(), contextKey{}, token))

		// Protect clients from caching the response
		rw.Header().Add(echo.HeaderVary, echo.HeaderCookie)
		h(rw, r)
	}
}

var tokenGenerators = map[string]func(*config, []string) (*tokenGenerator, error){
	"random": func(cfg *config, args []string) (*tokenGenerator, error) {
		var length uint8 = 32
		if len(args) > 0 {
			v, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return nil, err
			}
			length = uint8(v)
		}
		return &tokenGenerator{
			gen: func(r *http.Request) string {
				return random.String(length)
			},
			valid: func(cookieToken, clientToken string, r *http.Request) bool {
				return subtle.ConstantTimeCompare([]byte(cookieToken), []byte(clientToken)) == 1
			},
		}, nil
	},
}

// csrfTokenFromForm returns a `csrfTokenExtractor` that extracts token from the
// provided request header.
func csrfTokenFromHeader(header string) csrfTokenExtractor {
	return func(r *http.Request) (string, error) {
		return r.Header.Get(header), nil
	}
}

// csrfTokenFromForm returns a `csrfTokenExtractor` that extracts token from the
// provided form parameter.
func csrfTokenFromForm(param string) csrfTokenExtractor {
	return func(r *http.Request) (string, error) {
		token := r.FormValue(param)
		if token == "" {
			return "", errors.New("missing csrf token in the form parameter")
		}
		return token, nil
	}
}

// csrfTokenFromQuery returns a `csrfTokenExtractor` that extracts token from the
// provided query parameter.
func csrfTokenFromQuery(param string) csrfTokenExtractor {
	return func(r *http.Request) (string, error) {
		token := r.URL.Query().Get(param)
		if token == "" {
			return "", errors.New("missing csrf token in the query string")
		}
		return token, nil
	}
}

func init() {
	servicehub.Register("openapi-interceptor-csrf", &servicehub.Spec{
		Services:   []string{"openapi-interceptor-csrf"},
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
