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

package openapi

import (
	"context"
	"fmt"
	"net/http"

	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/providers/httpserver"
)

var _ Interface = (*service)(nil)

type service struct {
	p    *provider
	name string
}

var allMethods = []string{
	http.MethodConnect,
	http.MethodDelete,
	http.MethodGet,
	http.MethodHead,
	http.MethodOptions,
	http.MethodPatch,
	http.MethodPost,
	http.MethodPut,
	http.MethodTrace,
}

func (s *service) Add(method, path string, handler transhttp.HandlerFunc) {
	if len(method) <= 0 {
		for _, method := range allMethods {
			s.addRoute(method, path, handler)
		}
		return
	}
	s.addRoute(method, path, handler)
}

func (s *service) addRoute(method, path string, handler transhttp.HandlerFunc) {
	s.p.routes = append(s.p.routes, route{
		method:  method,
		path:    path,
		handler: s.p.WrapHandler(handler),
	})
}

func (s *service) WrapHandler(handler transhttp.HandlerFunc) transhttp.HandlerFunc {
	return s.p.WrapHandler(handler)
}

func (p *provider) WrapHandler(handler transhttp.HandlerFunc) transhttp.HandlerFunc {
	for i := len(p.interceptors) - 1; i >= 0; i-- {
		handler = transhttp.HandlerFunc(p.interceptors[i].Wrapper(http.HandlerFunc(handler)))
	}
	return handler
}

func (p *provider) Run(ctx context.Context) error {
	router := p.RouterManager.NewRouter()
	err := p.registerFixedRoutes(router)
	if err != nil {
		return err
	}
	err = router.Commit()
	if err != nil {
		return err
	}
	return p.doWatch(ctx)
}

func (p *provider) registerFixedRoutes(router httpserver.RouterTx) error {
	opts := []interface{}{httpserver.WithPathFormat(httpserver.PathFormatGoogleAPIs)}
	for _, route := range p.routes {
		err := router.Add(route.method, route.path, route.handler, opts...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *provider) doWatch(ctx context.Context) error {
	if len(p.watchers) > 0 && !p.RouterManager.Reloadable() {
		return fmt.Errorf("http router must be reloadable for route watchers")
	}
	for _, w := range p.watchers {
		go func(w RouteSourceWatcher) {
			ch := w.Watch()
			for {
				select {
				case source := <-ch:
					router := p.RouterManager.NewRouter()
					err := p.registerFixedRoutes(router)
					if err != nil {
						p.Log.Errorf("failed to load fixed routes: %s", err)
						continue
					}
					err = source.RegisterTo(&routerTx{
						tx: router,
						p:  p,
					})
					if err != nil {
						p.Log.Errorf("failed to register routes from source(%s): %s", source.Name(), err)
						continue
					}
					err = router.Commit()
					if err != nil {
						p.Log.Errorf("failed to commit routes from source(%s): %s", source.Name(), err)
					}
				case <-ctx.Done():
					return
				}
			}
		}(w)
	}
	return nil
}

type routerTx struct {
	tx httpserver.RouterTx
	p  *provider
}

func (r *routerTx) Add(method, path string, handler transhttp.HandlerFunc) {
	r.tx.Add(method, path, r.p.WrapHandler(handler), httpserver.WithPathFormat(httpserver.PathFormatGoogleAPIs))
}
