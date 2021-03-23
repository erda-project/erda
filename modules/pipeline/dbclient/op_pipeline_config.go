package dbclient

import (
	"github.com/erda-project/erda/modules/pipeline/spec"

	"github.com/mitchellh/mapstructure"
)

func (client *Client) ListPipelineConfigsOfActionExecutor() (configs []spec.PipelineConfig, cfgChan chan spec.ActionExecutorConfig, err error) {
	if err := client.Find(&configs, spec.PipelineConfig{Type: spec.PipelineConfigTypeActionExecutor}); err != nil {
		return nil, nil, err
	}
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
