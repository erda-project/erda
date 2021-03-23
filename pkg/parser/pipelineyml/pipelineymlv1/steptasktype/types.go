//go:generate stringer -type=StepTaskType
package steptasktype

type StepTaskType int

const (
	GET StepTaskType = iota
	PUT
	TASK
)
