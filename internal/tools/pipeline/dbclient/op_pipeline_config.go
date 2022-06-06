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

package dbclient

import (
	"github.com/mitchellh/mapstructure"

	spec2 "github.com/erda-project/erda/internal/tools/pipeline/spec"
)

// defaultAPITestActionExecutor provide api-test action-executor
var defaultAPITestActionExecutor = spec2.PipelineConfig{
	Type: spec2.PipelineConfigTypeActionExecutor,
	Value: spec2.ActionExecutorConfig{
		Kind:    string(spec2.PipelineTaskExecutorKindAPITest),
		Name:    spec2.PipelineTaskExecutorNameAPITestDefault.String(),
		Options: nil,
	},
}

var defaultWaitActionExecutor = spec2.PipelineConfig{
	Type: spec2.PipelineConfigTypeActionExecutor,
	Value: spec2.ActionExecutorConfig{
		Kind:    string(spec2.PipelineTaskExecutorKindWait),
		Name:    spec2.PipelineTaskExecutorNameWaitDefault.String(),
		Options: nil,
	},
}

var defaultK8sJobActionExecutor = spec2.PipelineConfig{
	Type: spec2.PipelineConfigTypeActionExecutor,
	Value: spec2.ActionExecutorConfig{
		Kind:    string(spec2.PipelineTaskExecutorKindK8sJob),
		Name:    spec2.PipelineTaskExecutorNameK8sJobDefault.String(),
		Options: nil,
	},
}

var defaultK8sFlinkActionExecutor = spec2.PipelineConfig{
	Type: spec2.PipelineConfigTypeActionExecutor,
	Value: spec2.ActionExecutorConfig{
		Kind:    string(spec2.PipelineTaskExecutorKindK8sFlink),
		Name:    spec2.PipelineTaskExecutorNameK8sFlinkDefault.String(),
		Options: nil,
	},
}

var defaultK8sSparkActionExecutor = spec2.PipelineConfig{
	Type: spec2.PipelineConfigTypeActionExecutor,
	Value: spec2.ActionExecutorConfig{
		Kind:    string(spec2.PipelineTaskExecutorKindK8sSpark),
		Name:    spec2.PipelineTaskExecutorNameK8sSparkDefault.String(),
		Options: nil,
	},
}

var defaultDockerActionExecutor = spec2.PipelineConfig{
	Type: spec2.PipelineConfigTypeActionExecutor,
	Value: spec2.ActionExecutorConfig{
		Kind:    string(spec2.PipelineTaskExecutorKindDocker),
		Name:    spec2.PipelineTaskExecutorNameDockerDefault.String(),
		Options: nil,
	},
}

func (client *Client) ListPipelineConfigsOfActionExecutor() (configs []spec2.PipelineConfig, cfgChan chan spec2.ActionExecutorConfig, err error) {
	if err := client.Find(&configs, spec2.PipelineConfig{Type: spec2.PipelineConfigTypeActionExecutor}); err != nil {
		return nil, nil, err
	}
	// add default api-test wait k8sjob k8sflink k8sspark action executor
	configs = append(configs, defaultAPITestActionExecutor, defaultWaitActionExecutor,
		defaultK8sJobActionExecutor, defaultK8sFlinkActionExecutor, defaultK8sSparkActionExecutor)
	cfgChan = make(chan spec2.ActionExecutorConfig, 100)
	for _, c := range configs {
		var r spec2.ActionExecutorConfig
		d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{Result: &r, ErrorUnused: true, TagName: "json"})
		if err != nil {
			return nil, nil, err
		}
		if err = d.Decode(c.Value); err != nil {
			return nil, nil, err
		}
		cfgChan <- r
	}
	return configs, cfgChan, nil
}
