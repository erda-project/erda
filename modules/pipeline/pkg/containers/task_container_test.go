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

package containers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
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
