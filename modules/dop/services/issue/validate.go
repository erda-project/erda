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
	iteration *dao.Iteration
	state     *dao.IssueState
	db        *dao.DBClient
}

type issueValidationConfig struct {
	iterationID int64
	stateID     int64
}

func NewIssueValidator(c *issueValidationConfig, db *dao.DBClient) (*issueValidator, error) {
	v := issueValidator{db: db}
	if c == nil {
		return &v, nil
	}
	if _, err := v.TryGetIteration(c.iterationID); err != nil {
		return nil, err
	}
	if _, err := v.TryGetState(c.stateID); err != nil {
		return nil, err
	}
	return &v, nil
}

func (v *issueValidator) TryGetIteration(iterationID int64) (*dao.Iteration, error) {
	if iterationID <= 0 {
		return nil, nil
	}
	if v.iteration != nil {
		return v.iteration, nil
	}
	iteration, err := v.db.GetIteration(uint64(iterationID))
	if err != nil {
		return nil, fmt.Errorf("failed to get iteration, err: %v", err)
	}
	v.iteration = iteration
	return iteration, nil
}

func (v *issueValidator) TryGetState(stateID int64) (*dao.IssueState, error) {
	if stateID <= 0 {
		return nil, nil
	}
	if v.state != nil {
		return v.state, nil
	}
	state, err := v.db.GetIssueStateByID(stateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue state, err: %v", err)
	}
	v.state = state
	return state, nil
}

func (v *issueValidator) validateIteration(time *time.Time) error {
	if time == nil || v.iteration == nil {
		return nil
	}
	if !v.inIterationInterval(time) {
		return fmt.Errorf("plan finish time is not in the iteration %v interval %v ~ %v",
			v.iteration.Title, v.iteration.StartedAt.Format("2006-01-02"), v.iteration.FinishedAt.Format("2006-01-02"))
	}
	return nil
}

func (v *issueValidator) inIterationInterval(time *time.Time) bool {
	if time == nil || v.iteration == nil {
		return false
	}
	return time.After(*v.iteration.StartedAt) && time.Before(*v.iteration.FinishedAt)
}

func (v *issueValidator) validateStateWithIteration() error {
	if v.iteration == nil {
		return nil
	}
	if v.iteration.State == apistructs.IterationStateFiled &&
		v.state.Belong != apistructs.IssueStateBelongDone && v.state.Belong != apistructs.IssueStateBelongClosed {
		return fmt.Errorf("put unfinished issue in archived iteration: %v is not allowed", v.iteration.Title)
	}
	return nil
}
