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

package diceyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var fieldname_validate_yml = `version: 2.0

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
    resource:			# err: resource
      cpu: 0.1
      mem: 256
    health_check:
      exec:
        cmds: echo 1		# err: cmds

  pmp-backend:
    deployments:
      replicas: 1
    ports:
    - 5080
    resources:
      cpu: 0.1
      mem: 512
      health_check:		# err: indent
          exec:
            cmd: echo 1

addons:
#  mysql:
#    plan: mysql:small
#    options:
#      create_dbs: pmp
  pmp-redis1:
    plans: redis:small		# err: plans
    image: redis:alpine
  pmp-zk:
    plan: zookeeper:medium
`

func TestFieldnameValidate(t *testing.T) {
	tests := []struct {
		name           string
		yamlData       []byte
		expectErrCount int
	}{
		{
			name:           "Valid YAML",
			yamlData:       []byte(fieldname_validate_yml),
			expectErrCount: 4,
		},
		{
			name:           "Invalid YAML",
			yamlData:       []byte(""),
			expectErrCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := New(tt.yamlData, false)
			assert.NoError(t, err)

			es := FieldnameValidate(d.Obj(), tt.yamlData)
			if len(es) != tt.expectErrCount {
				t.Fatalf("expect %d errors, but got %d errors", tt.expectErrCount, len(es))
			}
		})
	}
}
