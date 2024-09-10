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
	"bytes"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

var servicename_yml = `version: 2.0

envs:
  TERMINUS_APP_NAME: PMP
  TERMINUS_TRACE_ENABLE: true
  TRACE_SAMPLE: 1

services:
  pmp-ui-long-long-long:
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

  pmp_backend:
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

func TestServiceNameUnderline(t *testing.T) {
	d, err := New([]byte(servicename_yml), false)
	assert.Nil(t, err)
	es := ServiceNameCheck(d.Obj())
	assert.Equal(t, 1, len(es), "%v", es)
}

func TestEnvValidate(t *testing.T) {
	const baseYmlTemplate = `
envs:
{{- range $key, $value := .Envs }}
  {{ $key }}: '{{ $value }}'
{{- end }}
services:
  go-demo:
    deployments:
      replicas: 1
    image: go-demo-1725860282590763654
    ports:
    - 8080
    resources:
      cpu: 0.1
      mem: 128
{{- if .ServiceEnvs }}
    envs:
{{- range $key, $value := .ServiceEnvs }}
      {{ $key }}: '{{ $value }}'
{{- end }}
{{- end }}
version: "2.0"
`
	testCases := []struct {
		name        string
		envs        map[string]string
		serviceEnvs map[string]string
		expectErr   bool
	}{
		{
			name: "valid env with dot and dash",
			envs: map[string]string{
				"spring.cloud.compatibility-verifier.enabled": "false",
			},
			serviceEnvs: nil,
			expectErr:   false,
		},
		{
			name: "valid env with service level envs",
			envs: map[string]string{
				"spring.cloud.compatibility-verifier.enabled": "false",
			},
			serviceEnvs: map[string]string{
				"MY_SERVICE_ENV": "true",
			},
			expectErr: false,
		},
		{
			name: "invalid service level env with special characters",
			envs: map[string]string{
				"spring.cloud.compatibility-verifier.enabled": "false",
			},
			serviceEnvs: map[string]string{
				"my!service@env": "false",
			},
			expectErr: true,
		},
		{
			name: "valid service level env with numbers and dots",
			envs: map[string]string{
				"spring.cloud.compatibility-verifier.enabled": "false",
			},
			serviceEnvs: map[string]string{
				"service.env.123": "test",
			},
			expectErr: false,
		},
		{
			name: "valid start with number",
			envs: map[string]string{
				"8080port": "http",
			},
			expectErr: true,
		},
	}

	tmpl, err := template.New("yml").Parse(baseYmlTemplate)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var renderedYml bytes.Buffer

			err := tmpl.Execute(&renderedYml, map[string]interface{}{
				"Envs":        tc.envs,
				"ServiceEnvs": tc.serviceEnvs,
			})
			if err != nil {
				t.Fatalf("failed to render template: %v", err)
			}

			_, err = New(renderedYml.Bytes(), true)

			if tc.expectErr {
				assert.Error(t, err, "expected an error, but got none")
			} else {
				assert.NoError(t, err, "did not expect an error, but got one")
			}
		})
	}
}
