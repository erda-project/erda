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

package issueFilter

import (
	"encoding/json"
	"strconv"

	"github.com/erda-project/erda/apistructs"
)

type InParams struct {
	OrgID uint64 `json:"orgID,omitempty"`

	FrontendProjectID      string `json:"projectId,omitempty"`
	FrontendFixedIssueType string `json:"fixedIssueType,omitempty"`
	FrontendFixedIteration string `json:"fixedIteration,omitempty"`
	FrontendUrlQuery       string `json:"issueFilter__urlQuery,omitempty"`

	ProjectID   uint64                 `json:"-"`
	IssueTypes  []apistructs.IssueType `json:"-"`
	IterationID int64                  `json:"-"`
}

func (f *ComponentFilter) setInParams() error {
	b, err := json.Marshal(f.CtxBdl.InParams)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &f.InParams); err != nil {
		return err
	}

	f.InParams.OrgID, err = strconv.ParseUint(f.CtxBdl.Identity.OrgID, 10, 64)
	if err != nil {
		return err
	}

	// change type
	if f.InParams.FrontendProjectID != "" {
		f.InParams.ProjectID, err = strconv.ParseUint(f.InParams.FrontendProjectID, 10, 64)
		if err != nil {
			return err
		}
	}
	if f.InParams.FrontendFixedIssueType != "" {
		switch f.InParams.FrontendFixedIssueType {
		case "ALL":
			f.InParams.IssueTypes = []apistructs.IssueType{apistructs.IssueTypeEpic, apistructs.IssueTypeRequirement, apistructs.IssueTypeTask, apistructs.IssueTypeBug}
		case apistructs.IssueTypeEpic.String():
			f.InParams.IssueTypes = []apistructs.IssueType{apistructs.IssueTypeEpic}
		case apistructs.IssueTypeRequirement.String():
			f.InParams.IssueTypes = []apistructs.IssueType{apistructs.IssueTypeRequirement}
		case apistructs.IssueTypeTask.String():
			f.InParams.IssueTypes = []apistructs.IssueType{apistructs.IssueTypeTask}
		case apistructs.IssueTypeBug.String():
			f.InParams.IssueTypes = []apistructs.IssueType{apistructs.IssueTypeBug}
		}
	}
	if f.InParams.FrontendFixedIteration != "" {
		f.InParams.IterationID, err = strconv.ParseInt(f.InParams.FrontendFixedIteration, 10, 64)
		if err != nil {
			return err
		}
	}

	return nil
}
