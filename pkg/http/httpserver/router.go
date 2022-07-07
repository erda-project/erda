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
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/gorilla/mux"

	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

var registerOnce sync.Once

func (s *Server) RegisterToNewHttpServerRouter(newRouter httpserver.Router) error {
	duplicateRoutes, ok := s.CheckDuplicateRoutes()
	if !ok {
		for _, route := range duplicateRoutes {
			fmt.Printf("duplicate route: %s\n", route)
		}
		return fmt.Errorf("duplicate routes found")
	}
	registerOnce.Do(func() { newRouter.Any("/*", s.router) })
	return nil
}

func (s *Server) CheckDuplicateRoutes() ([]string, bool) {
	pathParamRegex := regexp.MustCompile(`{([^{}]*)}`)
	genKey := func(method, path string) string {
		path = strutil.ReplaceAllStringSubmatchFunc(pathParamRegex, path, func(i []string) string {
			return "{}"
		})
		return strings.ToLower(method) + ":" + path
	}
	// get all routes
	routeMap := make(map[string]int) // key: method + path
	_ = s.router.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		for _, method := range methods {
			key := genKey(method, path)
			routeMap[key]++
		}
		return nil
	})
	// check duplicate routes
	var duplicateRoutes []string
	for k, v := range routeMap {
		if v > 1 {
			duplicateRoutes = append(duplicateRoutes, k)
		}
	}
	return duplicateRoutes, len(duplicateRoutes) == 0
}
