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
