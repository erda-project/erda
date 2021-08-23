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
	"path/filepath"
	"regexp"
)

const (
	runtimeNameFormat = `^[a-zA-Z0-9\-]+$`
)

var (
	// record runtime's last restart time
	LastRestartTime = "lastRestartTime"
	//
	runtimeFormater *regexp.Regexp = regexp.MustCompile(runtimeNameFormat)
)

func makeRuntimeKey(namespace, name string) string {
	return filepath.Join("/dice/service/", namespace, name)
}

func validateRuntimeName(name string) bool {
	return len(name) > 0 && runtimeFormater.MatchString(name)
}

func validateRuntimeNamespace(namespace string) bool {
	return len(namespace) > 0 && runtimeFormater.MatchString(namespace)
}

func makePlatformKey(name string) string {
	return "/dice/service/platform/" + name
}
