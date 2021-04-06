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

package diceyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var mergeYml = `version: "2.0"
envs: {}
services:
    v-admin-center:
      ports:
      - 8080
      resources:
        cpu: 0.1
        max_cpu: 0.1
        mem: 256
      deployments:
        replicas: 1
      expose:
      - 80
      - 443
addons:
  monitor:
    plan: monitor:dev
environments:
  development: {}
  production:
    envs:
      DIFF_KEY: aaa
      TERMINUS_TRACE_ENABLE: "true"
    services:
      blog-web:
        envs:
          KEY: value
          KEY2: value2
          TERMINUS_APP_NAME: blog-web-jar
        resources:
          cpu: 0.1
          max_cpu: 0.1
          mem: 256
        deployments:
          replicas: 1
        health_check:
          http:
            port: 80
    addons:
      mysql:
        plan: rds
      redis:
        plan: redis:medium
  staging: {}
  test: {}
`

func TestMergeAddons(t *testing.T) {
	d, err := New([]byte(mergeYml), false)
	assert.Nil(t, err)
	MergeEnv(d.obj, "production")
	_, ok := d.obj.AddOns["mysql"]

	assert.True(t, ok)
	_, ok = d.obj.AddOns["redis"]
	assert.True(t, ok)
	_, ok = d.obj.AddOns["monitor"]
	assert.False(t, ok)
}
