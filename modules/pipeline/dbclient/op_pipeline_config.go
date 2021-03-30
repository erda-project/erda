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
		Name:    spec.PipelineTaskExecutorNameAPITestDefault,
		Options: nil,
	},
}

func (client *Client) ListPipelineConfigsOfActionExecutor() (configs []spec.PipelineConfig, cfgChan chan spec.ActionExecutorConfig, err error) {
	if err := client.Find(&configs, spec.PipelineConfig{Type: spec.PipelineConfigTypeActionExecutor}); err != nil {
		return nil, nil, err
	}
	// add default api-test action executor
	configs = append(configs, defaultAPITestActionExecutor)
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
