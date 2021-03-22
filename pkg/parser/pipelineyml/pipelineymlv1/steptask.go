package pipelineymlv1

import (
	"time"

	"github.com/erda-project/erda/pkg/parser/pipelineyml/pipelineymlv1/steptasktype"
)

type StepTask interface {
	Type() steptasktype.StepTaskType
	Name() string
	GetEnvs() map[string]string
	Validate() error
	OutputToContext() []string
	// includes required resources and explicit context
	RequiredContext(y PipelineYml) []string

	IsResourceTask() bool
	// if IsResourceTask, GetResourceType will return resource type name.
	GetResourceType() string

	IsDisable() bool
	IsPause() bool

	GetStatus() string
	SetStatus(status string)

	GetTaskParams() Params
	GetTaskConfig() TaskConfig
	GetCustomTaskRunPathArgs() []string
	GetResource() *Resource
	GetVersion() Version

	GetTimeout() (time.Duration, error)
}
