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

package containers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/pkg/parser/pipelineyml"

	"github.com/erda-project/erda/modules/pipeline/spec"
)

func TestGenContainers(t *testing.T) {
	k8sjobTask := &spec.PipelineTask{
		Name: "k8sjob",
		Extra: spec.PipelineTaskExtra{
			UUID: "pipeline-123456",
		},
	}
	k8sjobConainers, err := GenContainers(k8sjobTask)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(k8sjobConainers))

	k8sflinkTask := &spec.PipelineTask{
		Name: "k8sjob",
		Extra: spec.PipelineTaskExtra{
			UUID: "pipeline-123456",
			Action: pipelineyml.Action{
				Params: map[string]interface{}{
					"bigDataConf": "{\"args\":[],\"class\":\"io.terminus.dice.fdp.FlinkSqlMain\",\"flinkConf\":{\"jobManagerResource\":{\"cpu\":\"500.0m\",\"memory\":\"1024Mi\",\"replica\":1},\"kind\":\"FlinkJob\",\"parallelism\":1,\"taskManagerResource\":{\"cpu\":\"500.0m\",\"memory\":\"1024Mi\",\"replica\":1}},\"image\":\"registry.cn-hangzhou.aliyuncs.com/terminus/dice-flink:1.12.0\"}",
				},
			},
		},
	}
	k8sflinkContainers, err := GenContainers(k8sflinkTask)
	assert.Equal(t, nil, err)
	assert.Equal(t, 3, len(k8sflinkContainers))
}
