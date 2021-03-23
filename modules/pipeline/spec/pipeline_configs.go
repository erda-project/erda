package spec

type PipelineConfig struct {
	ID    uint64             `json:"id" xorm:"pk autoincr"`
	Type  PipelineConfigType `json:"type"`
	Value interface{}        `json:"value" xorm:"json"`
}

func (PipelineConfig) TableName() string {
	return "pipeline_configs"
}

type ActionExecutorConfig struct {
	Kind    string            `json:"kind,omitempty"`
	Name    string            `json:"name,omitempty"`
	Options map[string]string `json:"options,omitempty"`
}

type PipelineConfigType string

var (
	PipelineConfigTypeActionExecutor PipelineConfigType = "action_executor"
)
