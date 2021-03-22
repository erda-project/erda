package pipelineyml

import (
	"github.com/pkg/errors"
)

const (
	TimeoutDuration4Forever = -1
)

type TimeoutVisitor struct{}

func NewTimeoutVisitor() *TimeoutVisitor {
	return &TimeoutVisitor{}
}

func (v *TimeoutVisitor) Visit(s *Spec) {
	for stageIndex, stage := range s.Stages {
		for _, typedActionMap := range stage.Actions {
			for _, action := range typedActionMap {
				if action.Timeout < TimeoutDuration4Forever {
					s.appendError(errors.Errorf("invalid timeout: %d (only %d means forever)", action.Timeout, TimeoutDuration4Forever),
						stageIndex, action.Alias)
				}
			}
		}
	}
}
