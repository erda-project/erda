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

package precheck

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/precheck/prechecktype"
)

func TestPrecheck(t *testing.T) {
	ctx := prechecktype.InitContext()
	yamlByte := []byte(`
version: 1.1
stages:
- stage:
  - release:
      params:
        cross_cluster: false
`)
	items := prechecktype.ItemsForCheck{
		PipelineYml: "",
		Files: map[string]string{
			"dice.yml": `
version: "2.0"
services:
  java-demo:
    ports:
      - 8080
    expose:
      - 8080
    resources:
      cpu: 0.2
      mem: 512
    deployments:
      replicas: 1
`,
		},
		ActionSpecs: map[string]apistructs.ActionSpec{
			"release": {
				Params: []apistructs.ActionSpecParam{
					{
						Name:     "cross_cluster",
						Required: false,
						Default:  true,
					},
				},
			},
		},
	}
	_, _ = PreCheck(ctx, yamlByte, items)
	assert.False(t, prechecktype.GetContextResult(ctx, prechecktype.CtxResultKeyCrossCluster).(bool))
}

func TestPrecheckRelease(t *testing.T) {
	ctx := prechecktype.InitContext()
	yamlByte := []byte(`version: 1.1
stages:
- stage:
  - release:
      params:
        cross_cluster: false
`)
	items := prechecktype.ItemsForCheck{
		PipelineYml: "",
		Files: map[string]string{
			"dice.yml": `
version: 2
services:
  apm-demo-ui:
    testabc: abc
`,
		},
		ActionSpecs: map[string]apistructs.ActionSpec{
			"release": {
				Params: []apistructs.ActionSpecParam{
					{
						Name:     "cross_cluster",
						Required: false,
						Default:  true,
					},
				},
			},
		},
	}
	_, m := PreCheck(ctx, yamlByte, items)
	assert.Equal(t, m.Stacks[0], "taskName: release, message: failed to parse dice.yml, err: [\n  \"[apm-demo-ui] field 'testabc' not one of [image, cmd, ports, envs, hosts, labels, resources, volumes, deployments, depends_on, expose, health_check, binds, sidecars，init, traffic_security, endpoints, mesh_enable, k8s_snippet]\"\n]")
}
