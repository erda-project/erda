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
