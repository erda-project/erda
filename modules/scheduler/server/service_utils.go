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
