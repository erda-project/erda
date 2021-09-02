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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
)

func (svc *Issue) CreateIssueStage(req *apistructs.IssueStageRequest) error {
	err := svc.db.DeleteIssuesStage(req.OrgID, req.IssueType)
	if err != nil {
		return err
	}
	var stages []dao.IssueStage
	for _, v := range req.List {
		stage := dao.IssueStage{
			OrgID:     req.OrgID,
			IssueType: req.IssueType,
			Name:      v.Name,
			Value:     v.Value,
		}
		if stage.Value == "" {
			stage.Value = v.Name
		}
		stages = append(stages, stage)
	}
	return svc.db.CreateIssueStage(stages)
}

func (svc *Issue) GetIssueStage(req *apistructs.IssueStageRequest) ([]apistructs.IssueStage, error) {
	stages, err := svc.db.GetIssuesStage(req.OrgID, req.IssueType)
	if err != nil {
		return nil, err
	}
	var res []apistructs.IssueStage
	for _, v := range stages {
		stage := apistructs.IssueStage{
			ID:    int64(v.ID),
			Name:  v.Name,
			Value: v.Value,
		}
		if stage.Value == "" {
			stage.Value = v.Name
		}
		res = append(res, stage)
	}
	return res, nil
}
