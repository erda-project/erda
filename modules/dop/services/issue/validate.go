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

package issue

import (
	"fmt"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
)

type issueValidator struct {
}

type issueValidationConfig struct {
	iteration *dao.Iteration
	state     *dao.IssueState
}

func (v *issueValidator) validateTimeWithInIteration(c *issueValidationConfig, time *time.Time) error {
	if c == nil {
		return fmt.Errorf("issue validation config is empty")
	}
	if c.iteration == nil {
		return nil
	}
	if !inIterationInterval(c.iteration, time) {
		return fmt.Errorf("plan finish time is not in the iteration %v interval %v ~ %v",
			c.iteration.Title, c.iteration.StartedAt.Format("2006-01-02"), c.iteration.FinishedAt.Format("2006-01-02"))
	}
	return nil
}

func inIterationInterval(iteration *dao.Iteration, t *time.Time) bool {
	if t == nil || iteration == nil {
		return false
	}
	date := t.Truncate(24 * time.Hour)
	iterationStartAtDate, iterationEndAtDate := iteration.StartedAt.Truncate(24*time.Hour), iteration.FinishedAt.Truncate(24*time.Hour)
	return !date.Before(iterationStartAtDate) && !date.After(iterationEndAtDate)
}

func (v *issueValidator) validateStateWithIteration(c *issueValidationConfig) error {
	if c == nil {
		return fmt.Errorf("issue validation config is empty")
	}
	if c.iteration == nil {
		return nil
	}
	if c.iteration.State == apistructs.IterationStateFiled && (c.state == nil ||
		(c.state.Belong != apistructs.IssueStateBelongDone && c.state.Belong != apistructs.IssueStateBelongClosed)) {
		return fmt.Errorf("put unfinished issue in archived iteration: %v is not allowed", c.iteration.Title)
	}
	return nil
}

func (v *issueValidator) validateChangedFields(req *apistructs.IssueUpdateRequest, c *issueValidationConfig, changedFields map[string]interface{}) (err error) {
	if _, ok := changedFields["iteration_id"]; ok {
		if err = v.validateStateWithIteration(c); err != nil {
			return
		}
	} else {
		if _, ok := changedFields["plan_finished_at"]; ok {
			if err = v.validateTimeWithInIteration(c, req.PlanFinishedAt.Value()); err != nil {
				return
			}
		}
	}
	return
}
