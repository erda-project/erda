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

package oauth

import (
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

type referMatcher struct {
	allowedSuffixes []string
	exactDomains    map[string]struct{}
}

func (p *provider) buildReferMatcher() *referMatcher {
	r := referMatcher{
		allowedSuffixes: make([]string, 0),
		exactDomains: map[string]struct{}{
			p.Cfg.PlatformDomain: {},
			p.Cfg.IAMPublicURL:   {},
		},
	}
	for _, refer := range p.Cfg.AllowedReferrers {
		refer = strings.ToLower(strings.TrimSpace(refer))
		if strings.HasPrefix(refer, "*.") {
			// wildcard match: "*.test.com" => ".test.com"
			r.allowedSuffixes = append(r.allowedSuffixes, refer[1:]) // keep `.test.com`
		} else {
			r.exactDomains[refer] = struct{}{}
		}
	}

	logrus.Infof("allowed referrers: %v, %v", r.allowedSuffixes, r.exactDomains)
	return &r
}

func (r *referMatcher) Match(refer string) bool {
	if refer == "" {
		return true
	}

	u, err := url.Parse(refer)
	if err != nil {
		logrus.Warnf("illegal referer format %s, %v", refer, err)
		return false
	}

	host := strings.ToLower(strings.TrimSpace(u.Hostname()))

	// exact match
	if _, ok := r.exactDomains[host]; ok {
		return true
	}

	// suffix match
	for _, suffix := range r.allowedSuffixes {
		if !strings.HasSuffix(host, suffix) || len(host) == len(suffix) {
			continue
		}
		i := len(host) - len(suffix)
		if i >= 0 && host[i] == '.' {
			return true
		}
	}

	return false
}
