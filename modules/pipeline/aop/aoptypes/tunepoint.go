package aoptypes

// TuneType 调节的类型
type TuneType string

const (
	TuneTypePipeline TuneType = "pipeline" // pipeline 级别调节
	TuneTypeTask     TuneType = "task"     // task 级别调节
)

// TuneTrigger 调节的触发时机
type TuneTrigger string

const (
	TuneTriggerPipelineBeforeExec TuneTrigger = "pipeline_before_exec"
	TuneTriggerPipelineAfterExec  TuneTrigger = "pipeline_after_exec"

	TuneTriggerTaskBeforeExec    TuneTrigger = "task_before_exec"
	TuneTriggerTaskAfterExec     TuneTrigger = "task_after_exec"
	TuneTriggerTaskBeforePrepare TuneTrigger = "task_before_prepare"
	TuneTriggerTaskAfterPrepare  TuneTrigger = "task_after_prepare"
	TuneTriggerTaskBeforeCreate  TuneTrigger = "task_before_create"
	TuneTriggerTaskAfterCreate   TuneTrigger = "task_after_create"
	TuneTriggerTaskBeforeStart   TuneTrigger = "task_before_start"
	TuneTriggerTaskAfterStart    TuneTrigger = "task_after_start"
	TuneTriggerTaskBeforeQueue   TuneTrigger = "task_before_queue"
	TuneTriggerTaskAfterQueue    TuneTrigger = "task_after_queue"
	TuneTriggerTaskBeforeWait    TuneTrigger = "task_before_wait"
	TuneTriggerTaskAfterWait     TuneTrigger = "task_after_wait"
)

// TunePoint 调音点
type TunePoint interface {
	Type() TuneType
	Name() string
	Handle(TuneContext) error
}

type PipelineBaseTunePoint struct{}
type TaskBaseTunePoint struct{}

func (p PipelineBaseTunePoint) Type() TuneType { return TuneTypePipeline }
func (p TaskBaseTunePoint) Type() TuneType     { return TuneTypeTask }
