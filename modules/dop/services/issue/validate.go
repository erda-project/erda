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
	"github.com/erda-project/erda/pkg/cache"
)

type issueValidator struct {
	db             *dao.DBClient
	iterationCache *cache.Cache
	stateCache     *cache.Cache
}

type issueValidationConfig struct {
	iterationID int64
	stateID     int64
}

func NewIssueValidator(db *dao.DBClient) (*issueValidator, error) {
	v := issueValidator{db: db}
	v.iterationCache = cache.New("iteration", time.Minute, func(i interface{}) (interface{}, bool) {
		iteration, err := v.db.GetIteration(i.(uint64))
		if err != nil {
			return nil, false
		}
		return iteration, true
	})
	v.stateCache = cache.New("state", time.Minute, func(i interface{}) (interface{}, bool) {
		state, err := v.db.GetIssueStateByID(i.(int64))
		if err != nil {
			return nil, false
		}
		return state, true
	})
	return &v, nil
}

func (v *issueValidator) TryGetIteration(iterationID int64) (*dao.Iteration, error) {
	if iterationID <= 0 {
		return nil, nil
	}
	iteration, ok := v.iterationCache.LoadWithUpdate(uint64(iterationID))
	if !ok {
		return nil, fmt.Errorf("failed to get iteration")
	}
	return iteration.(*dao.Iteration), nil
}

func (v *issueValidator) TryGetState(stateID int64) (*dao.IssueState, error) {
	if stateID <= 0 {
		return nil, nil
	}
	state, ok := v.stateCache.LoadWithUpdate(stateID)
	if !ok {
		return nil, fmt.Errorf("failed to get state")
	}
	return state.(*dao.IssueState), nil
}

func (v *issueValidator) validateTimeWithInIteration(c *issueValidationConfig, time *time.Time) (*dao.Iteration, error) {
	if c == nil {
		return nil, fmt.Errorf("issue validation config is empty")
	}
	iteration, err := v.TryGetIteration(c.iterationID)
	if err != nil || iteration == nil {
		return nil, err
	}
	if !inIterationInterval(iteration, time) {
		return iteration, fmt.Errorf("plan finish time is not in the iteration %v interval %v ~ %v",
			iteration.Title, iteration.StartedAt.Format("2006-01-02"), iteration.FinishedAt.Format("2006-01-02"))
	}
	return iteration, nil
}

func inIterationInterval(iteration *dao.Iteration, time *time.Time) bool {
	if time == nil || iteration == nil {
		return false
	}
	return time.After(*iteration.StartedAt) && time.Before(*iteration.FinishedAt)
}

func (v *issueValidator) validateStateWithIteration(c *issueValidationConfig) error {
	if c == nil {
		return fmt.Errorf("issue validation config is empty")
	}
	iteration, err := v.TryGetIteration(c.iterationID)
	if err != nil {
		return err
	}
	state, err := v.TryGetState(c.stateID)
	if err != nil {
		return err
	}
	if iteration == nil {
		return nil
	}
	if iteration.State == apistructs.IterationStateFiled && (state == nil ||
		(state.Belong != apistructs.IssueStateBelongDone && state.Belong != apistructs.IssueStateBelongClosed)) {
		return fmt.Errorf("put unfinished issue in archived iteration: %v is not allowed", iteration.Title)
	}
	return nil
}
