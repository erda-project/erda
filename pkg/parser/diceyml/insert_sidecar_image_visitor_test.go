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

var yml2 = `version: 2.0

envs:
  TERMINUS_APP_NAME: PMP
  TERMINUS_TRACE_ENABLE: true
  TRACE_SAMPLE: 1

services:
  pmp-ui:
    ports:
    - 80
    expose:
    - 80
    deployments:
      replicas: 1
    depends_on:
    - pmp-backend
    resources:
      cpu: 0.1
      mem: 256
    health_check:
        exec:
          cmd: echo 1
    sidecars:
      xxx:
        cmd: ls
      yyy:
        cmd: ls

  pmp-backend:
    deployments:
      replicas: 1
    ports:
    - 5080
    resources:
      cpu: 0.1
      mem: 512
    health_check:
        exec:
          cmd: echo 1
    sidecars:
      xxx:
        cmd: ls

`

func TestInsertSideCarImage(t *testing.T) {
	d, err := New([]byte(yml2), false)
	assert.Nil(t, err)
	err = InsertSideCarImage(d.obj, map[string]map[string]string{
		"pmp-ui":      {"xxx": "xxx", "yyy": "xx"},
		"pmp-backend": {"xxx": "iii"},
	})
	assert.Nil(t, err)

}
