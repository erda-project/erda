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
