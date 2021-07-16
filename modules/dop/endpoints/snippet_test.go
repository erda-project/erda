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

package endpoints

import (
	"testing"

	"gotest.tools/assert"
)

func Test_getAppNameFromYmlPath(t *testing.T) {
	var tables = []struct {
		path    string
		appName string
	}{
		{
			path:    "name/dev/master/dice.yml",
			appName: "name",
		},
		{
			path:    "/baseName/dev/master/dice.yml",
			appName: "baseName",
		},
		{
			path:    "",
			appName: "",
		},
	}

	for _, v := range tables {
		assert.Equal(t, getAppNameFromYmlPath(v.path), v.appName)
	}
}

func Test_getBranchFromYmlPath(t *testing.T) {
	var tables = []struct {
		path   string
		branch string
		name   string
	}{
		{
			path:   "name/dev/master/dice.yml",
			branch: "master",
			name:   "/dice.yml",
		},
		{
			path:   "/baseName/dev/dev/aa/dice.yml",
			branch: "dev",
			name:   "/aa/dice.yml",
		},
		{
			path:   "64/DEV/feature/local/pipeline.yml",
			branch: "feature/local",
			name:   "/pipeline.yml",
		},
		{
			path:   "docker-spring-boot-java-web-service-example/TEST/develop/.dice/pipelines/xxx.yml",
			branch: "develop",
			name:   "/.dice/pipelines/xxx.yml",
		},
	}

	for _, v := range tables {
		assert.Equal(t, getBranchFromYmlPath(v.path, v.name), v.branch)
	}
}
