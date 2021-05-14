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

package server

import (
	"regexp"
)

const jobNameNamespaceFormat = `^[a-zA-Z0-9_\-]+$`

var jobFormater *regexp.Regexp = regexp.MustCompile(jobNameNamespaceFormat)

func validateJobName(name string) bool {
	return jobFormater.MatchString(name)
}

func validateJobNamespace(namespace string) bool {
	return jobFormater.MatchString(namespace)
}

func validateJobFlowID(id string) bool {
	return jobFormater.MatchString(id)
}

func makeJobKey(namespce, name string) string {
	return "/dice/job/" + namespce + "/" + name
}

func makeJobFlowKey(namespace, id string) string {
	return "/dice/jobflow/" + namespace + "/" + id
}
