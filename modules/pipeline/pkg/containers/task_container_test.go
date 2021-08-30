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
					"bigDataConf": "{\"args\":[\"--ACTION_OSS_ENDPOINT\",\"oss-cn-hangzhou.aliyuncs.com\",\"--ACTION_OSS_ACCESS_KEY_ID\",\"LTAI4GAPeSqzRDUNQ9EGGb55\",\"--ACTION_OSS_ACCESS_KEY_SECRET\",\"gUdhK5t2WdoC5RhPUgdJsMXpgncx6L\",\"--ACTION_OSS_BUCKET\",\"dice-files\",\"--ACTION_SCRIPT_PATH\",\"terminus-captain/6/6927/customer_CustomerBO_delta_sync1\",\"--ACTION_META_ADDR\",\"fdp-metadata-manager:8060\",\"--ACTION_META_DATA_BASE\",\"crm\"],\"class\":\"io.terminus.dice.fdp.FlinkSqlMain\",\"envs\":[{\"name\":\"test\",\"value\":\"test\"}],\"flinkConf\":{\"jobManagerResource\":{\"cpu\":\"500.0m\",\"memory\":\"1024Mi\",\"replica\":1},\"kind\":\"FlinkJob\",\"parallelism\":1,\"taskManagerResource\":{\"cpu\":\"500.0m\",\"memory\":\"1024Mi\",\"replica\":1}},\"image\":\"registry.cn-hangzhou.aliyuncs.com/terminus/dice-flink:1.12.0\",\"properties\":{\"classloader.resolve-order\":\"parent-first\",\"s3.access-key\":\"LTAI4GAPeSqzRDUNQ9EGGb55\",\"s3.endpoint\":\"oss-cn-hangzhou.aliyuncs.com\",\"s3.path.style.access\":\"true\",\"s3.secret-key\":\"gUdhK5t2WdoC5RhPUgdJsMXpgncx6L\",\"taskmanager.memory.flink.size\":\"1024M\",\"taskmanager.numberOfTaskSlots\":\"2\"},\"resource\":\"http://dice-files.oss-cn-hangzhou.aliyuncs.com/terminus-dev/flink/jar/fdp-flink-sql-20210630.jar\"}",
				},
			},
		},
	}
	k8sflinkContainers, err := GenContainers(k8sflinkTask)
	assert.Equal(t, nil, err)
	assert.Equal(t, 3, len(k8sflinkContainers))
}
