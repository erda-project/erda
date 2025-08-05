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

package router_define

import (
	"sort"
	"strings"
)

// RouteMatcher defines route matching interface
type RouteMatcher interface {
	MethodMatcher(method string) bool
	PathMatcher() PathMatcher
	GetPath() string
	GetMethod() string
}

// PathMatcher defines path matching interface
type PathMatcher interface {
	Match(path string) bool
}

// Router is responsible for managing and matching routes
type Router struct {
	routes []RouteMatcher
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		routes: make([]RouteMatcher, 0),
	}
}

// AddRoute adds a route
func (r *Router) AddRoute(route RouteMatcher) {
	r.routes = append(r.routes, route)
}

// FindBestMatch finds the best matching route
func (r *Router) FindBestMatch(method, path string) RouteMatcher {
	var candidates []RouteMatcher

	// 1. First find all matching routes
	for _, route := range r.routes {
		if route.MethodMatcher(method) && route.PathMatcher().Match(path) {
			candidates = append(candidates, route)
		}
	}

	// 2. If no matching routes, return nil
	if len(candidates) == 0 {
		return nil
	}

	// 3. Sort by priority, select highest priority
	sort.Slice(candidates, func(i, j int) bool {
		return CalculateRoutePriority(candidates[i]) > CalculateRoutePriority(candidates[j])
	})

	return candidates[0]
}

// CalculateRoutePriority calculates route priority (higher score means higher priority)
func CalculateRoutePriority(route RouteMatcher) int {
	score := 0
	path := route.GetPath()
	method := route.GetMethod()

	// 1. Path segment scoring
	segments := strings.Split(path, "/")
	validSegments := 0

	for _, segment := range segments {
		if segment != "" {
			validSegments++
			if strings.Contains(segment, "{") && strings.Contains(segment, "}") {
				// Parameter segments score lower
				score += 1
			} else {
				// Exact segments score higher
				score += 10
			}
		}
	}

	// 2. Path length bonus (longer paths are usually more precise)
	score += validSegments * 2

	// 3. Exact method matching bonus
	if method != "*" && method != "" && strings.ToUpper(method) != "ANY" {
		score += 5
	}

	// 4. Lower priority for root path
	if path == "/" {
		score -= 50
	}

	// 5. Lower priority for wildcard paths
	if strings.Contains(path, "*") {
		score -= 20
	}

	return score
}

// GetRoutes gets all routes (for testing)
func (r *Router) GetRoutes() []RouteMatcher {
	return append([]RouteMatcher(nil), r.routes...)
}

// Clear clears all routes (for testing)
func (r *Router) Clear() {
	r.routes = r.routes[:0]
}
