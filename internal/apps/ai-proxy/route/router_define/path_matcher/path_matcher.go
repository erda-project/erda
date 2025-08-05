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

package path_matcher

import (
	"regexp"
	"strings"
	"sync"
)

type PathMatcher struct {
	Pattern string
	Values  map[string]string

	match func(path string) bool
	mu    *sync.Mutex
}

func (p *PathMatcher) Match(path string) bool {
	return p.match(path)
}

func (p *PathMatcher) SetValue(key, value string) {
	p.mu.Lock()
	p.Values[key] = value
	p.mu.Unlock()
}

func (p *PathMatcher) GetValue(key string) (string, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	value, exists := p.Values[key]
	return value, exists
}

func NewPathMatcher(pattern string) *PathMatcher {
	pm := &PathMatcher{
		Pattern: pattern,
		Values:  make(map[string]string),
		mu:      &sync.Mutex{},
	}

	pm.match = func(path string) bool {
		quotedPattern := regexp.QuoteMeta(pattern)
		quotedPattern = strings.ReplaceAll(quotedPattern, `\{`, `(?P<`)
		quotedPattern = strings.ReplaceAll(quotedPattern, `\}`, `>[^/]+)`)
		re := regexp.MustCompile("^" + quotedPattern + "$")

		matches := re.FindStringSubmatch(path)
		if matches == nil {
			return false
		}

		pm.mu.Lock()
		defer pm.mu.Unlock()

		names := re.SubexpNames()
		for i, name := range names {
			if i > 0 && name != "" {
				pm.Values[name] = matches[i]
			}
		}
		return true
	}

	return pm
}
