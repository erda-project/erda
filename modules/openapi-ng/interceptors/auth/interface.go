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

package auth

import (
	"context"
	"net/http"
)

// AuthChecker .
type AuthChecker func(r *http.Request) (*CheckResult, error)

// CheckResult .
type CheckResult struct {
	Success bool
	Data    interface{}
}

// Auther .
type Auther interface {
	Name() string
	Match(r *http.Request) (AuthChecker, bool)
}

// AutherProvider .
type AutherProvider interface {
	Authers() []Auther
}

// AuthHandler .
type AuthHandler interface {
	RegisterHandler(add func(method, path string, h http.HandlerFunc))
}

type authInfoKey struct{}

// AuthInfo .
type AuthInfo struct {
	Type string
	Data interface{}
}

// WithAuthInfo .
func WithAuthInfo(ctx context.Context, info *AuthInfo) context.Context {
	return context.WithValue(ctx, authInfoKey{}, info)
}

// GetAuthInfo .
func GetAuthInfo(ctx context.Context) *AuthInfo {
	info, ok := ctx.Value(authInfoKey{}).(*AuthInfo)
	if !ok {
		return nil
	}
	return info
}

// AuthInfoSetter .
type AuthInfoSetter interface {
	SetAuthInfo(r *http.Request) *http.Request
}
