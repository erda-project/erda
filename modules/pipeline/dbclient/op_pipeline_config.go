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

package dbclient

import (
	"github.com/mitchellh/mapstructure"

	"github.com/erda-project/erda/modules/pipeline/spec"
)

// defaultAPITestActionExecutor provide api-test action-executor
var defaultAPITestActionExecutor = spec.PipelineConfig{
	Type: spec.PipelineConfigTypeActionExecutor,
	Value: spec.ActionExecutorConfig{
		Kind:    string(spec.PipelineTaskExecutorKindAPITest),
		Name:    spec.PipelineTaskExecutorNameAPITestDefault.String(),
		Options: nil,
	},
}

var defaultWaitActionExecutor = spec.PipelineConfig{
	Type: spec.PipelineConfigTypeActionExecutor,
	Value: spec.ActionExecutorConfig{
		Kind:    string(spec.PipelineTaskExecutorKindWait),
		Name:    spec.PipelineTaskExecutorNameWaitDefault.String(),
		Options: nil,
	},
}

func (client *Client) ListPipelineConfigsOfActionExecutor() (configs []spec.PipelineConfig, cfgChan chan spec.ActionExecutorConfig, err error) {
	if err := client.Find(&configs, spec.PipelineConfig{Type: spec.PipelineConfigTypeActionExecutor}); err != nil {
		return nil, nil, err
	}
	// add default api-test action executor
	configs = append(configs, defaultAPITestActionExecutor, defaultWaitActionExecutor)
	cfgChan = make(chan spec.ActionExecutorConfig, 100)
	for _, c := range configs {
		var r spec.ActionExecutorConfig
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
