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

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
)

func (svc *Issue) issueTimeIterationValidator(time *time.Time, iterationID int64) error {
	if iterationID == -1 {
		return nil
	}
	if time == nil {
		return errors.New("plan finish time is empty")
	}
	iteration, err := svc.db.GetIteration(uint64(iterationID))
	if err != nil {
		return fmt.Errorf("failed to get iteration, err: %v", err)
	}
	if time.Before(*iteration.StartedAt) || time.After(*iteration.FinishedAt) {
		return fmt.Errorf("plan finish time is not in the iteration  %v interval %v ~ %v", iteration.Title, iteration.StartedAt.Format("2006-01-02"), iteration.FinishedAt.Format("2006-01-02"))
	}
	return nil
}

func (svc *Issue) issueStateIterationValidator(stateID int64, iterationID int64) error {
	if iterationID == -1 {
		return nil
	}
	iteration, err := svc.db.GetIteration(uint64(iterationID))
	if err != nil {
		return fmt.Errorf("failed to get iteration, err: %v", err)
	}
	state, err := svc.db.GetIssueStateByID(stateID)
	if err != nil {
		return fmt.Errorf("failed to get issue state, err: %v", err)
	}
	if iteration.State == apistructs.IterationStateFiled && state.Belong != apistructs.IssueStateBelongDone && state.Belong != apistructs.IssueStateBelongClosed {
		return fmt.Errorf("put unfinished issue in archived iteration: %v is not allowed", iteration.Title)
	}
	return nil
}
