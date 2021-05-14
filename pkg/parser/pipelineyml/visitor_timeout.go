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
