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
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
)

func (svc *Issue) handleIssueStreamChangeIteration(lang i18n.LanguageCodes, currentIterationID, newIterationID int64) (
	streamType apistructs.IssueStreamType, params apistructs.ISTParam, err error) {

	// init default iteration
	unassignedIteration := &dao.Iteration{Title: svc.tran.Text(lang, "unassigned iteration")}
	currentIteration, newIteration := unassignedIteration, unassignedIteration
	streamType = apistructs.ISTChangeIteration

	// current iteration
	if currentIterationID == apistructs.UnassignedIterationID {
		streamType = apistructs.ISTChangeIterationFromUnassigned
	} else {
		currentIteration, err = svc.db.GetIteration(uint64(currentIterationID))
		if err != nil {
			return streamType, params, err
		}
	}

	// to iteration
	if newIterationID == apistructs.UnassignedIterationID {
		streamType = apistructs.ISTChangeIterationToUnassigned
	} else {
		newIteration, err = svc.db.GetIteration(uint64(newIterationID))
		if err != nil {
			return streamType, params, err
		}
	}

	params = apistructs.ISTParam{CurrentIteration: currentIteration.Title, NewIteration: newIteration.Title}

	return streamType, params, nil
}
