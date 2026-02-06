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

package auth

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/internal/core/user/auth/domain"
	"github.com/erda-project/erda/internal/core/user/auth/sessionrefresh"
	"github.com/erda-project/erda/internal/core/user/legacycontainer"
	"github.com/erda-project/erda/pkg/http/httputil"
)

type (
	// Interface .
	Interface interface {
		Interceptor(h http.HandlerFunc, opts func(r *http.Request) Options) http.HandlerFunc
	}
	// Auther .
	// Check returns (ok, req, user). When ok is true and user is non-nil, the framework calls ApplyUserInfoHeaders(req, user) once.
	Auther interface {
		Weight() int64
		Match(r *http.Request, opts Options) (bool, interface{})
		Check(r *http.Request, data interface{}, opts Options) (bool, *http.Request, error)
	}
	// Options .
	Options interface {
		Get(key string) interface{}
		Set(key string, val interface{})
	}
	// AutherLister .
	AutherLister interface {
		Authers() []Auther
	}
)

type provider struct {
	authers               []Auther
	overPermissionAuthers []Auther
}

func (p *provider) Init(ctx servicehub.Context) (err error) {
	ctx.Hub().ForeachServices(func(service string) bool {
		if strings.HasPrefix(service, "openapi-auth-") {
			authers, ok := ctx.Service(service).(AutherLister)
			if !ok {
				err = fmt.Errorf("%q not implements AutherLister", service)
				return false
			}
			p.authers = append(p.authers, authers.Authers()...)
		}

		if strings.HasPrefix(service, "openapi-over-permission-") {
			overPermissionAuthers, ok := ctx.Service(service).(AutherLister)
			if !ok {
				err = fmt.Errorf("%q not implements AutherLister", service)
				return false
			}
			p.overPermissionAuthers = append(p.overPermissionAuthers, overPermissionAuthers.Authers()...)
		}

		return true
	})
	if err != nil {
		return err
	}
	sort.Slice(p.authers, func(i, j int) bool {
		return p.authers[i].Weight() >= p.authers[j].Weight()
	})
	return nil
}

var _ Interface = (*provider)(nil)

func (p *provider) Interceptor(h http.HandlerFunc, opts func(r *http.Request) Options) http.HandlerFunc {
	if len(p.authers) <= 0 {
		return func(rw http.ResponseWriter, r *http.Request) {
			r.Header.Del(httputil.UserHeader)
			r.Header.Del(httputil.OrgHeader)
			h(rw, r)
		}
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		r.Header.Del(httputil.UserHeader)
		r.Header.Del(httputil.OrgHeader)
		opts := opts(r)
		if disable, _ := opts.Get("NoCheck").(bool); disable {
			h(rw, r)
			return
		}
		for _, auther := range p.authers {
			if ok, data := auther.Match(r, opts); ok {
				ok, req, err := auther.Check(r, data, opts)
				if err != nil {
					http.Error(rw, err.Error(), http.StatusUnauthorized)
					return
				}
				if ok {
					permissionAuthResult, err := p.executePermissionAuths(req, rw, opts)
					if err != nil {
						http.Error(rw, err.Error(), http.StatusUnauthorized)
						return
					}
					if !permissionAuthResult {
						break // execute error
					}
					if refresh := sessionrefresh.Get(req.Context()); refresh != nil {
						if writer := legacycontainer.Get[domain.RefreshWriter](); writer != nil {
							if err := writer.WriteRefresh(rw, req, refresh); err != nil {
								logrus.Warnf("failed to write session refresh: %v", err)
							}
						}
					}
					h(rw, req)
					return
				}
				break

			}
		}
		http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}
}

func (p *provider) executePermissionAuths(r *http.Request, rw http.ResponseWriter, opts Options) (bool, error) {
	if len(p.overPermissionAuthers) == 0 {
		return true, nil
	}
	for _, auth := range p.overPermissionAuthers {
		if ok, data := auth.Match(r, opts); ok {
			ok, _, err := auth.Check(r, data, opts)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
		}
	}
	return true, nil
}

func init() {
	servicehub.Register("openapi-auth", &servicehub.Spec{
		Services: []string{"openapi-auth"},
		DependenciesFunc: func(hub *servicehub.Hub) (list []string) {
			hub.ForeachServices(func(service string) bool {
				if strings.HasPrefix(service, "openapi-auth-") {
					list = append(list, service)
				}
				return true
			})
			return list
		},
		Creator: func() servicehub.Provider { return &provider{} },
	})
}
