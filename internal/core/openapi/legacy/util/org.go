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

package util

import (
	"fmt"
	"strings"

	"github.com/erda-project/erda/internal/core/openapi/legacy/conf"
	"github.com/erda-project/erda/pkg/strutil"
)

func GetOrgByDomain(domain string) (string, error) {
	valid := false
	for _, rootDomain := range conf.RootDomainList() {
		if strings.HasSuffix(domain, rootDomain) {
			valid = true
			break
		}
	}
	if !valid {
		return "", fmt.Errorf("invalid domain")
	}
	for _, rootDomain := range conf.RootDomainList() {
		if orgName := orgNameRetriever(domain, rootDomain); orgName != "" {
			return orgName, nil
		}
	}
	return "", nil
}

func orgNameRetriever(domain, rootDomain string) string {
	suf := strutil.Concat(".", rootDomain)
	domain_and_port := strutil.Split(domain, ":", true)
	if len(domain_and_port) == 0 {
		return ""
	}
	domain = domain_and_port[0]
	if strutil.HasSuffixes(domain, suf) {
		orgName := strutil.TrimSuffixes(domain, suf)
		if strutil.HasSuffixes(orgName, "-org") {
			orgName = strutil.TrimSuffixes(orgName, "-org")
		}
		return orgName
	}
	return ""
}
