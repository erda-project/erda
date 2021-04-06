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

var compatible_expose_yml = `version: 2.0
services:
  web:
    ports:
      - 8080
      - 20880
      - port: 1234
        protocol: "UDP"
      - port: 12345
        protocol: "TCP"
        expose: true
    health_check:
      exec:
        cmd: "echo 1"
    deployments:
      replicas: 1
    resources:
      cpu: 0.1
      mem: 512
      disk: 0
    expose:
      - 20880
      - 1234
    volumes:
      - storage: "nfs"
        path: "/data/file/resource"`

func TestCompatibleExpose(t *testing.T) {
	d, err := New([]byte(compatible_expose_yml), false) // new的时候已经执行了CompatibleExpose
	assert.Nil(t, err)
	assert.True(t, d.obj.Services["web"].Ports[0].Expose)
}
