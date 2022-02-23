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

package container_provider

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

func TestConstructContainerProvider(t *testing.T) {
	validLabel := map[string]string{
		apistructs.ContainerInstanceLabelType:     string(apistructs.ContainerInstanceECI),
		apistructs.ContainerInstanceLabelCPU:      "1.0",
		apistructs.ContainerInstanceLabelMemoryMB: "1024",
	}
	validProvider := ConstructContainerProvider(WithLabels(validLabel), WithStages([]*pipelineyml.Stage{}))
	assert.Equal(t, apistructs.ContainerInstanceECI, validProvider.ContainerInstanceType)

	invalidLabel := map[string]string{
		apistructs.ContainerInstanceLabelType:     "xxx",
		apistructs.ContainerInstanceLabelCPU:      "1.0",
		apistructs.ContainerInstanceLabelMemoryMB: "1024",
	}
	invalidProvider := ConstructContainerProvider(WithLabels(invalidLabel), WithStages([]*pipelineyml.Stage{}))
	assert.Equal(t, false, invalidProvider.IsHitted)

	validTypeButInvalidCPU := map[string]string{
		apistructs.ContainerInstanceLabelType:     string(apistructs.ContainerInstanceECI),
		apistructs.ContainerInstanceLabelCPU:      "xxx",
		apistructs.ContainerInstanceLabelMemoryMB: "1024",
	}
	validTypeButInvalidCPUProvider := ConstructContainerProvider(WithLabels(validTypeButInvalidCPU), WithStages([]*pipelineyml.Stage{}))
	assert.Equal(t, true, validTypeButInvalidCPUProvider.IsHitted)
	assert.Equal(t, float64(0), validTypeButInvalidCPUProvider.CPU)

	validTypeButInvalidMem := map[string]string{
		apistructs.ContainerInstanceLabelType:     string(apistructs.ContainerInstanceECI),
		apistructs.ContainerInstanceLabelCPU:      "1.0",
		apistructs.ContainerInstanceLabelMemoryMB: "xxx",
	}
	validTypeButInvalidMemProvider := ConstructContainerProvider(WithLabels(validTypeButInvalidMem), WithStages([]*pipelineyml.Stage{}))
	assert.Equal(t, true, validTypeButInvalidMemProvider.IsHitted)
	assert.Equal(t, float64(0), validTypeButInvalidMemProvider.MemoryMB)

	customPipelineYml := `version: "1.1"
stages:
  - stage:
      - custom-script:
          alias: custom-script
          description: 运行自定义命令
          version: "1.0"
          commands:
            - sleep 1200
            - sleep 21
          resources:
            cpu: 0.5
            mem: 1024`
	pipelineYml, err := pipelineyml.New([]byte(customPipelineYml))
	assert.NoError(t, err)
	customPipelineProvider := ConstructContainerProvider(WithLabels(validLabel), WithStages(pipelineYml.Spec().Stages))
	assert.Equal(t, true, customPipelineProvider.IsDisabled)
	assert.Equal(t, false, customPipelineProvider.IsHitted)
}

func TestDealPipelineProviderBeforeRun(t *testing.T) {
	p := &spec.Pipeline{
		PipelineBase: spec.PipelineBase{
			ID: 1,
		},
		PipelineExtra: spec.PipelineExtra{
			Extra: spec.PipelineExtraInfo{
				ContainerInstanceProvider: nil,
			},
		},
	}
	clusterInfo := apistructs.ClusterInfoData{
		apistructs.ECIEnable:  "true",
		apistructs.ECIHitRate: "100",
	}
	DealPipelineProviderBeforeRun(p, clusterInfo)
	assert.Equal(t, true, p.Extra.ContainerInstanceProvider.IsHitted)
	assert.Equal(t, apistructs.ContainerInstanceECI, p.Extra.ContainerInstanceProvider.ContainerInstanceType)
}

func TestDealJobAndClusterInfo(t *testing.T) {
	job := &apistructs.JobFromUser{
		ContainerInstanceProvider: &apistructs.ContainerInstanceProvider{
			IsHitted:              true,
			ContainerInstanceType: apistructs.ContainerInstanceECI,
		},
		Env: map[string]string{},
	}
	clusterInfo := map[string]string{
		apistructs.ECIEnable:  "false",
		apistructs.ECIHitRate: "0",
	}
	DealJobAndClusterInfo(job, clusterInfo)
	assert.Equal(t, "true", clusterInfo[apistructs.BuildkitEnable])
	assert.Equal(t, "100", clusterInfo[apistructs.BuildkitHitRate])
	assert.Equal(t, "true", job.Env["ACTIONAGENT_ENABLE_PUSH_LOG_TO_COLLECTOR"])
}
